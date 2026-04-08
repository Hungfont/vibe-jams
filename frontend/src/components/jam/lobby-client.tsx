"use client";

import * as React from "react";
import { useRouter } from "next/navigation";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ToastStack } from "@/components/ui/toast";
import { useToast } from "@/components/ui/use-toast";
import type { ApiEnvelope } from "@/lib/api/types";
import { joinJamSchema } from "@/lib/jam/schemas";
import type { SessionSnapshot } from "@/lib/jam/types";

export function LobbyClient() {
  const router = useRouter();
  const { items, toast } = useToast();
  const [mode, setMode] = React.useState("create");
  const [jamId, setJamId] = React.useState("");
  const [isSubmitting, setIsSubmitting] = React.useState(false);
  const [errorText, setErrorText] = React.useState<string | null>(null);

  const parseEnvelope = React.useCallback(async <T,>(response: Response): Promise<ApiEnvelope<T>> => {
    return (await response.json()) as ApiEnvelope<T>;
  }, []);

  const runCreate = React.useCallback(async () => {
    setIsSubmitting(true);
    setErrorText(null);

    const response = await fetch("/api/jam/create", { method: "POST" });
    const envelope = await parseEnvelope<SessionSnapshot>(response);
    if (!envelope.success || !envelope.data) {
      const message = envelope.error?.message ?? "Unable to create Jam";
      setErrorText(message);
      toast({ title: "Create failed", description: message, variant: "error" });
      setIsSubmitting(false);
      return;
    }

    toast({ title: "Jam created", description: `Session ${envelope.data.jamId}` });
    router.push(`/jam/${encodeURIComponent(envelope.data.jamId)}`);
  }, [parseEnvelope, router, toast]);

  const runJoin = React.useCallback(async () => {
    setIsSubmitting(true);
    setErrorText(null);

    const parsed = joinJamSchema.safeParse({ jamId });
    if (!parsed.success) {
      setErrorText("Jam ID is invalid");
      setIsSubmitting(false);
      return;
    }

    const response = await fetch(`/api/jam/${encodeURIComponent(parsed.data.jamId)}/join`, {
      method: "POST",
    });
    const envelope = await parseEnvelope<SessionSnapshot>(response);
    if (!envelope.success || !envelope.data) {
      const message = envelope.error?.message ?? "Unable to join Jam";
      setErrorText(message);
      toast({ title: "Join failed", description: message, variant: "error" });
      setIsSubmitting(false);
      return;
    }

    toast({ title: "Joined Jam", description: `Session ${envelope.data.jamId}` });
    router.push(`/jam/${encodeURIComponent(envelope.data.jamId)}`);
  }, [jamId, parseEnvelope, router, toast]);

  const onSubmit = React.useCallback(async () => {
    if (mode === "create") {
      await runCreate();
    } else {
      await runJoin();
    }
    setIsSubmitting(false);
  }, [mode, runCreate, runJoin]);

  return (
    <main className="relative flex min-h-screen items-center justify-center bg-gradient-to-b from-zinc-900 via-black to-black px-6 py-10">
      <Card className="w-full max-w-lg border-zinc-800 bg-zinc-950/90">
        <CardHeader>
          <CardTitle className="text-xl">Jam Lobby</CardTitle>
          <CardDescription>
            Start or join a collaborative listening room inspired by Spotify Jam.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {errorText ? (
            <Alert className="border-rose-500/50 bg-rose-950/40">
              <AlertTitle>Action failed</AlertTitle>
              <AlertDescription>{errorText}</AlertDescription>
            </Alert>
          ) : null}

          <Tabs defaultValue="create" value={mode} onValueChange={setMode}>
            <TabsList>
              <TabsTrigger value="create">Create Jam</TabsTrigger>
              <TabsTrigger value="join">Join Jam</TabsTrigger>
            </TabsList>

            <TabsContent value="create" className="space-y-2">
              <p className="text-xs text-zinc-400">Premium users can host and control playback.</p>
            </TabsContent>

            <TabsContent value="join" className="space-y-2">
              <Input
                value={jamId}
                onChange={(event) => setJamId(event.target.value)}
                placeholder="Enter jamId"
              />
            </TabsContent>
          </Tabs>

          <Button className="w-full" disabled={isSubmitting} onClick={() => void onSubmit()}>
            {isSubmitting ? "Processing..." : mode === "create" ? "Create Jam" : "Join Jam"}
          </Button>
        </CardContent>
      </Card>

      <ToastStack items={items} />
    </main>
  );
}
