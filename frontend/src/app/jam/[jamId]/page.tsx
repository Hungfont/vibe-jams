import { JamRoomClient } from "@/components/jam/jam-room-client";
import { ROOM_TABS, type RoomTab } from "@/lib/jam/constants";
import type { BffOrchestrationData } from "@/lib/jam/types";
import { getAppBaseUrl, getRequestAuthHeaders } from "@/lib/api/server";

export const dynamic = "force-dynamic";

interface JamPageProps {
  params: Promise<{ jamId: string }>;
  searchParams: Promise<{ view?: string }>;
}

async function getInitialData(jamId: string): Promise<{
  initialData: BffOrchestrationData | null;
  initialError?: { code: string; message: string };
}> {
  const baseUrl = await getAppBaseUrl();
  const auth = await getRequestAuthHeaders();

  const response = await fetch(`${baseUrl}/api/bff/jam/${encodeURIComponent(jamId)}/orchestration`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...(auth.authorization ? { Authorization: auth.authorization } : {}),
      ...(auth.cookie ? { Cookie: auth.cookie } : {}),
    },
    cache: "no-store",
  });

  const payload = (await response.json()) as {
    success: boolean;
    data?: BffOrchestrationData;
    error?: { code: string; message: string };
  };

  if (!payload.success || !payload.data) {
    return {
      initialData: null,
      initialError: payload.error ?? { code: "internal_error", message: "Failed to load room" },
    };
  }

  return { initialData: payload.data };
}

function resolveView(raw: string | undefined): RoomTab {
  if (!raw) {
    return "queue";
  }

  const value = raw.trim().toLowerCase();
  if ((ROOM_TABS as readonly string[]).includes(value)) {
    return value as RoomTab;
  }

  return "queue";
}

export default async function JamPage({ params, searchParams }: JamPageProps) {
  const { jamId } = await params;
  const query = await searchParams;
  const initialView = resolveView(query.view);
  const { initialData, initialError } = await getInitialData(jamId);

  return (
    <JamRoomClient
      jamId={jamId}
      initialData={initialData}
      initialError={initialError}
      initialView={initialView}
    />
  );
}
