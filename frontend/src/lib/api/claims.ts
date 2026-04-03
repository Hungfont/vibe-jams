import { backendJson } from "@/lib/api/http";
import type { Claims } from "@/lib/jam/types";

const PREMIUM_PLANS = new Set(["premium", "premium_plus", "pro"]);

export interface ClaimValidationResult {
  ok: boolean;
  claims?: Claims;
  code?: string;
  message?: string;
}

export async function validateClaims(authHeader: string, cookieHeader: string): Promise<ClaimValidationResult> {
  const result = await backendJson<Claims>({
    service: "auth",
    path: "/internal/v1/auth/validate",
    method: "POST",
    authHeader,
    cookieHeader,
  });

  if (!result.ok || !result.data) {
    return {
      ok: false,
      code: result.error?.code ?? "unauthorized",
      message: result.error?.message ?? "unauthorized",
    };
  }

  return {
    ok: true,
    claims: result.data,
  };
}

export function isPremiumPlan(plan: string): boolean {
  return PREMIUM_PLANS.has(plan.trim().toLowerCase());
}
