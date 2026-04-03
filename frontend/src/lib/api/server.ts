import { headers } from "next/headers";

export async function getAppBaseUrl(): Promise<string> {
  const requestHeaders = await headers();
  const host = requestHeaders.get("host") ?? "localhost:3000";
  const protocol = requestHeaders.get("x-forwarded-proto") ?? "http";
  return `${protocol}://${host}`;
}

export async function getRequestAuthHeaders(): Promise<{ authorization?: string; cookie?: string }> {
  const requestHeaders = await headers();
  const authorization = requestHeaders.get("authorization") ?? undefined;
  const cookie = requestHeaders.get("cookie") ?? undefined;
  return { authorization, cookie };
}
