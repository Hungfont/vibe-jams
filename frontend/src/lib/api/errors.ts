import type { ApiErrorBody } from "@/lib/api/types";

function asRecord(value: unknown): Record<string, unknown> | null {
  if (!value || typeof value !== "object") {
    return null;
  }
  return value as Record<string, unknown>;
}

export function extractErrorBody(payload: unknown): Partial<ApiErrorBody> {
  const root = asRecord(payload);
  if (!root) {
    return {};
  }

  const nestedError = asRecord(root.error);
  if (nestedError) {
    return {
      code: typeof nestedError.code === "string" ? nestedError.code : undefined,
      message: typeof nestedError.message === "string" ? nestedError.message : undefined,
      dependency: typeof nestedError.dependency === "string" ? nestedError.dependency : undefined,
    };
  }

  if (root.success === false) {
    const flatError = asRecord(root.error);
    if (flatError) {
      return {
        code: typeof flatError.code === "string" ? flatError.code : undefined,
        message: typeof flatError.message === "string" ? flatError.message : undefined,
        dependency: typeof flatError.dependency === "string" ? flatError.dependency : undefined,
      };
    }
  }

  return {
    code: typeof root.code === "string" ? root.code : undefined,
    message: typeof root.message === "string" ? root.message : undefined,
    dependency: typeof root.dependency === "string" ? root.dependency : undefined,
  };
}

export function mapHttpStatusToDefaultCode(status: number): string {
  if (status === 400) {
    return "invalid_input";
  }
  if (status === 401) {
    return "unauthorized";
  }
  if (status === 403) {
    return "forbidden";
  }
  if (status === 404) {
    return "not_found";
  }
  if (status === 409) {
    return "version_conflict";
  }
  if (status === 503) {
    return "dependency_unavailable";
  }
  return "internal_error";
}

export function normalizeUpstreamError(status: number, payload: unknown, fallbackMessage: string): ApiErrorBody {
  const extracted = extractErrorBody(payload);
  return {
    code: extracted.code ?? mapHttpStatusToDefaultCode(status),
    message: extracted.message ?? fallbackMessage,
    dependency: extracted.dependency,
  };
}
