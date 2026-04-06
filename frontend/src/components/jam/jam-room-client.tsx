"use client";

import * as React from "react";
import useSWR from "swr";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Avatar } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { Slider } from "@/components/ui/slider";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ToastStack } from "@/components/ui/toast";
import { Tooltip } from "@/components/ui/tooltip";
import { useToast } from "@/components/ui/use-toast";
import type { ApiEnvelope } from "@/lib/api/types";
import { addQueueItem, kickParticipant, muteParticipant, removeQueueItem, reorderQueue } from "@/lib/jam/actions";
import { endJamSession, executePlayback, fetchJamState, fetchOrchestration, fetchRealtimeWsConfig } from "@/lib/jam/client";
import { BLOCKING_CODES, type RoomTab } from "@/lib/jam/constants";
import { isSnapshotRecovery, reduceRealtimeVersion, type RoomRealtimeEvent } from "@/lib/jam/realtime";
import type { BffOrchestrationData, QueueSnapshot, SessionSnapshot } from "@/lib/jam/types";

interface JamRoomClientProps {
  jamId: string;
  initialView: RoomTab;
  initialData: BffOrchestrationData | null;
  initialError?: { code: string; message: string };
}

const ORCHESTRATION_REFRESH_MS = 20000;

async function loadRoom(jamId: string): Promise<BffOrchestrationData> {
  const envelope = await fetchOrchestration(jamId);
  if (!envelope.success || !envelope.data) {
    throw new Error(envelope.error?.message ?? "Failed to load room");
  }
  return envelope.data;
}

function mergeQueue(room: BffOrchestrationData, snapshot: QueueSnapshot): BffOrchestrationData {
  return {
    ...room,
    sessionState: {
      ...room.sessionState,
      queue: snapshot,
      aggregateVersion: Math.max(room.sessionState.aggregateVersion, snapshot.queueVersion),
    },
  };
}

export function JamRoomClient({ jamId, initialView, initialData, initialError }: JamRoomClientProps) {
  const { items, toast } = useToast();
  const [trackId, setTrackId] = React.useState("");
  const [activeTab, setActiveTab] = React.useState<RoomTab>(initialView);
  const [dialogOpen, setDialogOpen] = React.useState(false);
  const [positionMs, setPositionMs] = React.useState(0);
  const [moderationReason, setModerationReason] = React.useState("");
  const [localError, setLocalError] = React.useState(initialError?.message ?? "");
  const [connectionState, setConnectionState] = React.useState("disconnected");
  const [pendingPlayback, setPendingPlayback] = React.useState<string | null>(null);

  const swr = useSWR(["room", jamId], () => loadRoom(jamId), {
    refreshInterval: ORCHESTRATION_REFRESH_MS,
    revalidateOnFocus: false,
    fallbackData: initialData ?? undefined,
  });

  const room = swr.data;
  const isEnded = room?.sessionState.session.status === "ended";
  const isHost = room ? room.claims.userId === room.sessionState.session.hostUserId : false;

  const refreshSnapshot = React.useCallback(async () => {
    const state = await fetchJamState(jamId);
    if (!state.success || !state.data) {
      setLocalError(state.error?.message ?? "Failed to refresh snapshot");
      return;
    }
    const snapshot = state.data;

    swr.mutate((current) => {
      if (!current) {
        return current;
      }
      return {
        ...current,
        sessionState: snapshot,
      };
    }, false);
    setPendingPlayback(null);
  }, [jamId, swr]);

  const runQueueMutation = React.useCallback(
    async (mutation: () => Promise<ApiEnvelope<QueueSnapshot>>): Promise<ApiEnvelope<QueueSnapshot>> => {
      let response = await mutation();
      if (response.success || response.error?.code !== "version_conflict") {
        return response;
      }

      await refreshSnapshot();
      response = await mutation();
      return response;
    },
    [refreshSnapshot],
  );

  React.useEffect(() => {
    let socket: WebSocket | null = null;
    let mounted = true;

    async function connect() {
      setConnectionState("connecting");
      const lastSeenVersion = room?.sessionState.aggregateVersion ?? 0;
      const bootstrap = await fetchRealtimeWsConfig(jamId, lastSeenVersion);

      if (!bootstrap.success || !bootstrap.data || !mounted) {
        setConnectionState("degraded");
        return;
      }

      const query = `sessionId=${encodeURIComponent(bootstrap.data.sessionId)}&lastSeenVersion=${encodeURIComponent(bootstrap.data.lastSeenVersion)}`;
      socket = new WebSocket(`${bootstrap.data.wsUrl}?${query}`);

      socket.onopen = () => setConnectionState("connected");
      socket.onclose = () => setConnectionState("disconnected");
      socket.onerror = () => setConnectionState("degraded");
      socket.onmessage = (event) => {
        const parsed = JSON.parse(event.data) as RoomRealtimeEvent;

        const currentVersion = swr.data?.sessionState.aggregateVersion ?? 0;
        if (isSnapshotRecovery(parsed)) {
          const snapshot = parsed.payload;
          swr.mutate(
            (current) => {
              if (!current) {
                return current;
              }
              return {
                ...current,
                sessionState: snapshot,
              };
            },
            false,
          );
          toast({ title: "Recovered", description: "Room synchronized from snapshot" });
          return;
        }

        const decision = reduceRealtimeVersion(currentVersion, parsed);
        if (decision === "apply") {
          void refreshSnapshot();
          return;
        }

        if (decision === "recover") {
          void refreshSnapshot();
        }
      };
    }

    void connect();

    return () => {
      mounted = false;
      if (socket) {
        socket.close();
      }
    };
  }, [jamId, refreshSnapshot, room?.sessionState.aggregateVersion, swr, toast]);

  const handleAddTrack = React.useCallback(async () => {
    if (!room || !trackId || isEnded) {
      return;
    }

    const payload = {
      trackId,
      addedBy: room.claims.userId,
      idempotencyKey: `${Date.now()}-${Math.random()}`,
    };

    const response = await runQueueMutation(() => addQueueItem(jamId, payload));

    if (!response.success || !response.data) {
      setLocalError(response.error?.message ?? "Unable to add track");
      toast({ title: "Add failed", description: response.error?.message, variant: "error" });
      return;
    }

    swr.mutate((current) => (current ? mergeQueue(current, response.data as QueueSnapshot) : current), false);
    setTrackId("");
    toast({ title: "Track added", description: response.data.items.at(-1)?.trackId ?? "Queue updated" });
  }, [room, trackId, isEnded, jamId, swr, toast, runQueueMutation]);

  const handleRemove = React.useCallback(
    async (itemId: string) => {
      const response = await runQueueMutation(() => removeQueueItem(jamId, { itemId }));
      if (!response.success || !response.data) {
        setLocalError(response.error?.message ?? "Unable to remove item");
        toast({ title: "Remove failed", description: response.error?.message, variant: "error" });
        return;
      }
      swr.mutate((current) => (current ? mergeQueue(current, response.data as QueueSnapshot) : current), false);
    },
    [jamId, runQueueMutation, swr, toast],
  );

  const handleReorderReverse = React.useCallback(async () => {
    if (!room || isEnded) {
      return;
    }

    const reversed = [...room.sessionState.queue.items].map((item) => item.itemId).reverse();
    const payload = {
      itemIds: reversed,
      expectedQueueVersion: room.sessionState.queue.queueVersion,
    };
    const response = await runQueueMutation(() => reorderQueue(jamId, payload));

    if (!response.success || !response.data) {
      setLocalError(response.error?.message ?? "Unable to reorder queue");
      toast({ title: "Reorder failed", description: response.error?.message, variant: "error" });
      return;
    }

    swr.mutate((current) => (current ? mergeQueue(current, response.data as QueueSnapshot) : current), false);
  }, [isEnded, jamId, room, runQueueMutation, swr, toast]);

  const handlePlayback = React.useCallback(
    async (command: "play" | "pause" | "next" | "prev" | "seek") => {
      if (!room || isEnded || !isHost || pendingPlayback) {
        return;
      }

      const response = await executePlayback(jamId, {
        command,
        clientEventId: `${Date.now()}-${Math.random()}`,
        expectedQueueVersion: room.sessionState.queue.queueVersion,
        positionMs: command === "seek" ? positionMs : undefined,
      });

      if (!response.success) {
        setLocalError(response.error?.message ?? "Playback command rejected");
        toast({ title: "Playback rejected", description: response.error?.message, variant: "error" });
        setPendingPlayback(null);
        if (response.error?.code === "version_conflict") {
          await refreshSnapshot();
        }
        return;
      }

      setPendingPlayback(command);
      toast({ title: "Playback accepted", description: `Command ${command} queued` });
    },
    [isEnded, isHost, jamId, pendingPlayback, positionMs, refreshSnapshot, room, toast],
  );

  React.useEffect(() => {
    if (!pendingPlayback) {
      return;
    }

    const timeout = setTimeout(() => {
      setPendingPlayback(null);
    }, 2500);

    return () => clearTimeout(timeout);
  }, [pendingPlayback]);

  const endSession = React.useCallback(async () => {
    const envelope = await endJamSession(jamId);

    if (!envelope.success) {
      setLocalError(envelope.error?.message ?? "Unable to end session");
      toast({ title: "End failed", description: envelope.error?.message, variant: "error" });
      return;
    }

    setDialogOpen(false);
    await refreshSnapshot();
  }, [jamId, refreshSnapshot, toast]);

  const handleModeration = React.useCallback(
    async (action: "mute" | "kick", targetUserId: string) => {
      if (!room || !isHost || isEnded) {
        return;
      }

      const payload = {
        targetUserId,
        reason: moderationReason.trim() ? moderationReason.trim() : undefined,
      };
      const response: ApiEnvelope<SessionSnapshot> =
        action === "mute"
          ? await muteParticipant(jamId, payload)
          : await kickParticipant(jamId, payload);

      if (!response.success || !response.data) {
        setLocalError(response.error?.message ?? `Unable to ${action} participant`);
        toast({ title: `${action} failed`, description: response.error?.message, variant: "error" });
        return;
      }
      const sessionSnapshot = response.data;

      swr.mutate(
        (current) => {
          if (!current) {
            return current;
          }
          return {
            ...current,
            sessionState: {
              ...current.sessionState,
              session: sessionSnapshot,
              aggregateVersion: Math.max(current.sessionState.aggregateVersion, sessionSnapshot.sessionVersion),
            },
          };
        },
        false,
      );
      toast({ title: `Participant ${action}d`, description: `${targetUserId} updated` });
    },
    [isEnded, isHost, jamId, moderationReason, room, swr, toast],
  );

  if (swr.isLoading && !room) {
    return (
      <main className="min-h-screen bg-black p-6">
        <Skeleton className="mb-4 h-8 w-56" />
        <Skeleton className="mb-2 h-52 w-full" />
        <Skeleton className="h-52 w-full" />
      </main>
    );
  }

  if (!room) {
    return (
      <main className="min-h-screen bg-black p-6">
        <Alert className="border-rose-500/60 bg-rose-950/40">
          <AlertTitle>Room unavailable</AlertTitle>
          <AlertDescription>{initialError?.message ?? "Unable to load room state."}</AlertDescription>
        </Alert>
      </main>
    );
  }

  const blockingError = room.issues?.find((issue) => BLOCKING_CODES.has(issue.code));

  return (
    <main className="min-h-screen bg-black text-zinc-100">
      <div className="grid min-h-screen grid-cols-1 lg:grid-cols-[220px_1fr]">
        <aside className="border-r border-zinc-900 p-4">
          <p className="text-lg font-semibold">Jam</p>
          <p className="text-xs text-zinc-400">Session {room.sessionState.session.jamId}</p>
          <div className="mt-4 space-y-2">
            <Badge>{room.claims.plan}</Badge>
            <Badge>{connectionState}</Badge>
            <Badge>{room.sessionState.session.status}</Badge>
          </div>
          <Separator className="my-4" />
          <Button variant="ghost" className="w-full justify-start" onClick={() => setActiveTab("queue")}>Queue</Button>
          <Button variant="ghost" className="w-full justify-start" onClick={() => setActiveTab("playback")}>Playback</Button>
          <Button variant="ghost" className="w-full justify-start" onClick={() => setActiveTab("participants")}>Participants</Button>
          <Button variant="ghost" className="w-full justify-start" onClick={() => setActiveTab("diagnostics")}>Diagnostics</Button>
        </aside>

        <section className="flex min-h-screen flex-col">
          <header className="sticky top-0 z-20 border-b border-zinc-900 bg-zinc-950/90 p-4 backdrop-blur">
            <div className="flex items-center justify-between gap-4">
              <div>
                <h1 className="text-xl font-semibold">Jam Room</h1>
                <p className="text-xs text-zinc-400">Host: {room.sessionState.session.hostUserId}</p>
              </div>
              <div className="flex items-center gap-2">
                <Badge>Participants {room.sessionState.session.participants.length}</Badge>
                <Badge>Queue v{room.sessionState.queue.queueVersion}</Badge>
              </div>
            </div>
          </header>

          <div className="flex-1 space-y-4 p-4 pb-28">
            {blockingError || localError ? (
              <Alert className="border-rose-500/60 bg-rose-950/40">
                <AlertTitle>{blockingError ? blockingError.code : "Action error"}</AlertTitle>
                <AlertDescription>{blockingError?.message ?? localError}</AlertDescription>
              </Alert>
            ) : null}

            {room.partial ? (
              <Alert>
                <AlertTitle>Dependency degraded</AlertTitle>
                <AlertDescription>
                  Optional dependencies are degraded. Core room controls remain available where possible.
                </AlertDescription>
              </Alert>
            ) : null}

            <Tabs defaultValue={initialView} value={activeTab} onValueChange={(value) => setActiveTab(value as RoomTab)}>
              <TabsList>
                <TabsTrigger value="queue">Queue</TabsTrigger>
                <TabsTrigger value="playback">Playback</TabsTrigger>
                <TabsTrigger value="participants">Participants</TabsTrigger>
                <TabsTrigger value="diagnostics">Diagnostics</TabsTrigger>
              </TabsList>

              <TabsContent value="queue">
                <Card>
                  <CardHeader>
                    <CardTitle>Queue</CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-3">
                    <div className="flex gap-2">
                      <Input
                        value={trackId}
                        onChange={(event) => setTrackId(event.target.value)}
                        disabled={isEnded}
                        placeholder="Track ID"
                      />
                      <Button onClick={() => void handleAddTrack()} disabled={isEnded || !trackId}>Add</Button>
                      <Button variant="secondary" onClick={() => void handleReorderReverse()} disabled={isEnded}>
                        Reverse
                      </Button>
                    </div>
                    <ScrollArea>
                      <div className="space-y-2">
                        {room.sessionState.queue.items.map((item) => (
                          <Card key={item.itemId} className="border-zinc-800 bg-zinc-900">
                            <CardContent className="flex items-center justify-between p-3">
                              <div>
                                <p className="text-sm font-medium">{item.trackId}</p>
                                <p className="text-xs text-zinc-400">added by {item.addedBy}</p>
                              </div>
                              <DropdownMenu>
                                <DropdownMenuTrigger asChild>
                                  <Button variant="outline" size="sm">Actions</Button>
                                </DropdownMenuTrigger>
                                <DropdownMenuContent>
                                  <DropdownMenuItem onClick={() => void handleRemove(item.itemId)}>Remove</DropdownMenuItem>
                                </DropdownMenuContent>
                              </DropdownMenu>
                            </CardContent>
                          </Card>
                        ))}
                      </div>
                    </ScrollArea>
                  </CardContent>
                </Card>
              </TabsContent>

              <TabsContent value="playback">
                <Card>
                  <CardHeader>
                    <CardTitle>Playback</CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="grid grid-cols-2 gap-2 sm:grid-cols-5">
                      {(["play", "pause", "next", "prev", "seek"] as const).map((command) => (
                        <Tooltip
                          key={command}
                          content={!isHost ? "Host only control" : isEnded ? "Session ended" : command}
                        >
                          <Button
                            className="w-full"
                            variant={command === "pause" ? "secondary" : "default"}
                            disabled={!isHost || isEnded || Boolean(pendingPlayback)}
                            onClick={() => void handlePlayback(command)}
                          >
                            {command}
                          </Button>
                        </Tooltip>
                      ))}
                    </div>
                    <div className="space-y-2">
                      <p className="text-xs text-zinc-400">Seek position (ms)</p>
                      <Slider
                        min={0}
                        max={300000}
                        value={positionMs}
                        onChange={(event) => setPositionMs(Number(event.target.value))}
                        disabled={!isHost || isEnded || Boolean(pendingPlayback)}
                      />
                    </div>
                    {pendingPlayback ? <p className="text-xs text-amber-300">Pending command: {pendingPlayback}</p> : null}
                  </CardContent>
                </Card>
              </TabsContent>

              <TabsContent value="participants">
                <Card>
                  <CardHeader>
                    <CardTitle>Participants</CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-3">
                    {isHost ? (
                      <Input
                        value={moderationReason}
                        onChange={(event) => setModerationReason(event.target.value)}
                        placeholder="Moderation reason (optional)"
                        disabled={isEnded}
                      />
                    ) : null}
                    {room.sessionState.session.participants.map((participant) => (
                      <div
                        className="flex items-center justify-between rounded-md border border-zinc-800 bg-zinc-900 px-3 py-2"
                        key={`${participant.userId}-${participant.role}`}
                      >
                        <div className="flex items-center gap-2">
                          <Avatar fallback={participant.userId.slice(0, 2).toUpperCase()} />
                          <span className="text-sm">{participant.userId}</span>
                        </div>
                        <div className="flex items-center gap-2">
                          {participant.muted ? <Badge className="border-amber-700 bg-amber-950 text-amber-200">muted</Badge> : null}
                          <Badge>{participant.role}</Badge>
                          {isHost && participant.userId !== room.sessionState.session.hostUserId ? (
                            <>
                              <Button
                                size="sm"
                                variant="outline"
                                disabled={isEnded || Boolean(participant.muted)}
                                onClick={() => void handleModeration("mute", participant.userId)}
                              >
                                Mute
                              </Button>
                              <Button
                                size="sm"
                                variant="destructive"
                                disabled={isEnded}
                                onClick={() => void handleModeration("kick", participant.userId)}
                              >
                                Kick
                              </Button>
                            </>
                          ) : null}
                        </div>
                      </div>
                    ))}
                  </CardContent>
                </Card>
              </TabsContent>

              <TabsContent value="diagnostics">
                <Card>
                  <CardHeader>
                    <CardTitle>Diagnostics</CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-2 text-sm">
                    <p className="text-zinc-300">Connection: {connectionState}</p>
                    <p className="text-zinc-300">Partial: {String(room.partial)}</p>
                    <p className="text-zinc-300">Aggregate version: {room.sessionState.aggregateVersion}</p>
                    {room.issues?.map((issue) => (
                      <Alert key={`${issue.dependency}-${issue.code}`}>
                        <AlertTitle>{issue.dependency}</AlertTitle>
                        <AlertDescription>
                          {issue.code}: {issue.message}
                        </AlertDescription>
                      </Alert>
                    ))}
                  </CardContent>
                </Card>
              </TabsContent>
            </Tabs>
          </div>

          <footer className="fixed inset-x-0 bottom-0 border-t border-zinc-900 bg-zinc-950/95 p-3">
            <div className="mx-auto flex max-w-6xl items-center justify-between gap-3">
              <div>
                <p className="text-xs text-zinc-400">Now Playing</p>
                <p className="text-sm font-medium">Queue items: {room.sessionState.queue.items.length}</p>
              </div>
              <div className="flex items-center gap-2">
                <Button size="sm" variant="secondary" disabled={!isHost || isEnded || Boolean(pendingPlayback)} onClick={() => void handlePlayback("prev")}>Prev</Button>
                <Button size="sm" disabled={!isHost || isEnded || Boolean(pendingPlayback)} onClick={() => void handlePlayback("play")}>Play</Button>
                <Button size="sm" variant="secondary" disabled={!isHost || isEnded || Boolean(pendingPlayback)} onClick={() => void handlePlayback("next")}>Next</Button>
              </div>
              <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
                <DialogTrigger asChild>
                  <Button size="sm" variant="destructive" disabled={!isHost || isEnded}>End Session</Button>
                </DialogTrigger>
                <DialogContent>
                  <DialogHeader>
                    <DialogTitle>End this Jam?</DialogTitle>
                    <DialogDescription>
                      This action ends the room for all participants and blocks further writes.
                    </DialogDescription>
                  </DialogHeader>
                  <DialogFooter>
                    <DialogClose asChild>
                      <Button variant="outline" size="sm">Cancel</Button>
                    </DialogClose>
                    <Button variant="destructive" size="sm" onClick={() => void endSession()}>
                      End session
                    </Button>
                  </DialogFooter>
                </DialogContent>
              </Dialog>
            </div>
          </footer>
        </section>
      </div>

      <ToastStack items={items} />
    </main>
  );
}
