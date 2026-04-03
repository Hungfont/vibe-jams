import type { ApiEnvelope } from "@/lib/api/types";
import type { QueueSnapshot } from "@/lib/jam/types";
import {
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
