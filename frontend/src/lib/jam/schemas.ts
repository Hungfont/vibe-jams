import { z } from "zod";

export const joinJamSchema = z.object({
  jamId: z.string().trim().min(3, "jamId is required"),
});

export const queueAddSchema = z.object({
  trackId: z.string().trim().min(1, "trackId is required"),
  addedBy: z.string().trim().min(1, "addedBy is required"),
  idempotencyKey: z.string().trim().min(1, "idempotencyKey is required"),
});

export const queueRemoveSchema = z.object({
  itemId: z.string().trim().min(1, "itemId is required"),
});

export const queueReorderSchema = z.object({
  itemIds: z.array(z.string().trim().min(1)).min(1, "itemIds is required"),
  expectedQueueVersion: z.number().int().positive("expectedQueueVersion must be positive"),
});

export const playbackCommandSchema = z.object({
  command: z.enum(["play", "pause", "next", "prev", "seek"]),
  trackId: z.string().trim().optional(),
  clientEventId: z.string().trim().min(1, "clientEventId is required"),
  expectedQueueVersion: z.number().int().positive("expectedQueueVersion must be positive"),
  positionMs: z.number().int().min(0).optional(),
});
