import { z } from "zod";

export const claimsSchema = z.object({
  userId: z.string().trim().min(1, "userId is required"),
  plan: z.string().trim().min(1, "plan is required"),
  sessionState: z.string().trim().min(1, "sessionState is required"),
  scope: z.array(z.string().trim().min(1)).optional(),
});

export const loginRequestSchema = z.object({
  identity: z.string().trim().email("Enter a valid email"),
  password: z
    .string()
    .min(8, "Password must be at least 8 characters")
    .max(128, "Password is too long"),
});

export const refreshRequestSchema = z.object({
  refreshToken: z.string().trim().min(1).optional(),
});

export const logoutRequestSchema = z.object({
  refreshToken: z.string().trim().min(1).optional(),
});

export const authTokenPairSchema = z.object({
  accessToken: z.string().trim().min(1, "accessToken is required"),
  refreshToken: z.string().trim().min(1, "refreshToken is required"),
  tokenType: z.string().trim().min(1),
  expiresAt: z.string().trim().min(1, "expiresAt is required"),
  claims: claimsSchema,
});

export type LoginRequest = z.infer<typeof loginRequestSchema>;
export type ClaimsPayload = z.infer<typeof claimsSchema>;