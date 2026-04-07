import type { ApiEnvelope } from "@/lib/api/types";
import type { QueueSnapshot, SessionPermissions, SessionSnapshot } from "@/lib/jam/types";
import {
  moderationCommandSchema,
  permissionUpdateSchema,
  queueAddSchema,
  queueRemoveSchema,
  queueReorderSchema,
} from "@/lib/jam/schemas";

export async function addQueueItem(jamId: string, payload: unknown): Promise<ApiEnvelope<QueueSnapshot>> {
  const parsed = queueAddSchema.safeParse(payload);
  if (!parsed.success) {
    return {
      success: false,
      error: {
        code: "invalid_input",
        message: "invalid add payload",
      },
    };
  }

  const response = await fetch(`/api/jam/${encodeURIComponent(jamId)}/queue/add`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(parsed.data),
  });

  return (await response.json()) as ApiEnvelope<QueueSnapshot>;
}

export async function removeQueueItem(jamId: string, payload: unknown): Promise<ApiEnvelope<QueueSnapshot>> {
  const parsed = queueRemoveSchema.safeParse(payload);
  if (!parsed.success) {
    return {
      success: false,
      error: {
        code: "invalid_input",
        message: "invalid remove payload",
      },
    };
  }

  const response = await fetch(`/api/jam/${encodeURIComponent(jamId)}/queue/remove`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(parsed.data),
  });

  return (await response.json()) as ApiEnvelope<QueueSnapshot>;
}

export async function reorderQueue(jamId: string, payload: unknown): Promise<ApiEnvelope<QueueSnapshot>> {
  const parsed = queueReorderSchema.safeParse(payload);
  if (!parsed.success) {
    return {
      success: false,
      error: {
        code: "invalid_input",
        message: "invalid reorder payload",
      },
    };
  }

  const response = await fetch(`/api/jam/${encodeURIComponent(jamId)}/queue/reorder`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(parsed.data),
  });

  return (await response.json()) as ApiEnvelope<QueueSnapshot>;
}

export async function muteParticipant(jamId: string, payload: unknown): Promise<ApiEnvelope<SessionSnapshot>> {
  const parsed = moderationCommandSchema.safeParse(payload);
  if (!parsed.success) {
    return {
      success: false,
      error: {
        code: "invalid_input",
        message: "invalid mute payload",
      },
    };
  }

  const response = await fetch(`/api/jam/${encodeURIComponent(jamId)}/moderation/mute`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(parsed.data),
  });

  return (await response.json()) as ApiEnvelope<SessionSnapshot>;
}

export async function kickParticipant(jamId: string, payload: unknown): Promise<ApiEnvelope<SessionSnapshot>> {
  const parsed = moderationCommandSchema.safeParse(payload);
  if (!parsed.success) {
    return {
      success: false,
      error: {
        code: "invalid_input",
        message: "invalid kick payload",
      },
    };
  }

  const response = await fetch(`/api/jam/${encodeURIComponent(jamId)}/moderation/kick`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(parsed.data),
  });

  return (await response.json()) as ApiEnvelope<SessionSnapshot>;
}

export async function updateSessionPermissions(
  jamId: string,
  payload: unknown,
): Promise<ApiEnvelope<SessionPermissions>> {
  const parsed = permissionUpdateSchema.safeParse(payload);
  if (!parsed.success) {
    return {
      success: false,
      error: {
        code: "invalid_input",
        message: "invalid permissions payload",
      },
    };
  }

  if (
    parsed.data.canControlPlayback === undefined &&
    parsed.data.canReorderQueue === undefined &&
    parsed.data.canChangeVolume === undefined
  ) {
    return {
      success: false,
      error: {
        code: "invalid_input",
        message: "at least one permission field is required",
      },
    };
  }

  const response = await fetch(`/api/jam/${encodeURIComponent(jamId)}/permissions`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(parsed.data),
  });

  return (await response.json()) as ApiEnvelope<SessionPermissions>;
}
