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
  expectedQueueVersion: z.number().int().positive("expectedQueueVersion must be positive"),
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

export const moderationCommandSchema = z.object({
  targetUserId: z.string().trim().min(1, "targetUserId is required"),
  reason: z.string().trim().optional(),
});

export const permissionUpdateSchema = z.object({
  canControlPlayback: z.boolean().optional(),
  canReorderQueue: z.boolean().optional(),
  canChangeVolume: z.boolean().optional(),
});

const claimsSchema = z.object({
  userId: z.string().trim().min(1, "userId is required"),
  plan: z.string().trim().min(1, "plan is required"),
  sessionState: z.string().trim().min(1, "sessionState is required"),
  scope: z.array(z.string().trim().min(1)).optional(),
});

const sessionParticipantSchema = z.object({
  userId: z.string().trim().min(1, "participant userId is required"),
  role: z.string().trim().min(1, "participant role is required"),
  muted: z.boolean().optional(),
});

const sessionSnapshotSchema = z.object({
  jamId: z.string().trim().min(1, "jamId is required"),
  status: z.string().trim().min(1, "status is required"),
  hostUserId: z.string().trim().min(1, "hostUserId is required"),
  participants: z.array(sessionParticipantSchema),
  permissions: z
    .object({
      canControlPlayback: z.boolean(),
      canReorderQueue: z.boolean(),
      canChangeVolume: z.boolean(),
    })
    .optional(),
  sessionVersion: z.number().int().nonnegative("sessionVersion must be non-negative"),
  endCause: z.string().trim().optional(),
  endedBy: z.string().trim().optional(),
});

const queueItemSchema = z.object({
  itemId: z.string().trim().min(1, "itemId is required"),
  trackId: z.string().trim().min(1, "trackId is required"),
  addedBy: z.string().trim().min(1, "addedBy is required"),
});

const queueSnapshotSchema = z.object({
  jamId: z.string().trim().min(1, "jamId is required"),
  queueVersion: z.number().int().nonnegative("queueVersion must be non-negative"),
  items: z.array(queueItemSchema),
});

const sessionStateSnapshotSchema = z.object({
  session: sessionSnapshotSchema,
  queue: queueSnapshotSchema,
  aggregateVersion: z.number().int().nonnegative("aggregateVersion must be non-negative"),
});

const bffIssueSchema = z.object({
  dependency: z.string().trim().min(1, "dependency is required"),
  code: z.string().trim().min(1, "code is required"),
  message: z.string().trim().min(1, "message is required"),
});

const trackLookupSchema = z.object({
  trackId: z.string().trim().min(1, "trackId is required"),
  isPlayable: z.boolean(),
  reasonCode: z.string().trim().optional(),
  title: z.string().trim().optional(),
  artist: z.string().trim().optional(),
});

export const bffOrchestrationDataSchema = z.object({
  claims: claimsSchema,
  sessionState: sessionStateSnapshotSchema,
  track: trackLookupSchema.optional(),
  partial: z.boolean(),
  dependencyStatuses: z.record(z.string(), z.string()),
  issues: z.array(bffIssueSchema).optional(),
});

const bffErrorSchema = z.object({
  code: z.string().trim().min(1),
  message: z.string().trim().min(1),
  dependency: z.string().trim().optional(),
});

export const bffUpstreamEnvelopeSchema = z.discriminatedUnion("success", [
  z.object({
    success: z.literal(true),
    data: bffOrchestrationDataSchema,
  }),
  z.object({
    success: z.literal(false),
    error: bffErrorSchema.optional(),
  }),
]);
