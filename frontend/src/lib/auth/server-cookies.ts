import type { NextRequest, NextResponse } from "next/server";

export const ACCESS_TOKEN_COOKIE = "auth_token";
export const REFRESH_TOKEN_COOKIE = "refresh_token";
export const CSRF_TOKEN_COOKIE = "csrf_token";

const DEFAULT_ACCESS_TTL_SECONDS = 10 * 60;
const DEFAULT_REFRESH_TTL_SECONDS = 7 * 24 * 60 * 60;

function isSecureCookie(): boolean {
  return process.env.NODE_ENV === "production";
}

function parseAccessTokenTTL(expiresAt: string): number {
  const timestamp = Date.parse(expiresAt);
  if (Number.isNaN(timestamp)) {
    return DEFAULT_ACCESS_TTL_SECONDS;
  }

  const now = Date.now();
  const seconds = Math.floor((timestamp - now) / 1000);
  if (seconds <= 0) {
    return DEFAULT_ACCESS_TTL_SECONDS;
  }

  return seconds;
}

export function issueCsrfToken(): string {
  if (typeof crypto.randomUUID === "function") {
    return crypto.randomUUID().replaceAll("-", "");
  }

  const random = Math.random().toString(36).slice(2);
  return `${Date.now()}${random}`;
}

export function readRefreshToken(request: NextRequest): string {
  return request.cookies.get(REFRESH_TOKEN_COOKIE)?.value?.trim() ?? "";
}

export function isValidCsrfRequest(request: NextRequest): boolean {
  const header = request.headers.get("x-csrf-token")?.trim() ?? "";
  const cookie = request.cookies.get(CSRF_TOKEN_COOKIE)?.value?.trim() ?? "";

  return header.length > 0 && cookie.length > 0 && header === cookie;
}

export function setAuthCookies(
  response: NextResponse,
  accessToken: string,
  refreshToken: string,
  expiresAt: string,
  csrfToken: string,
): void {
  response.cookies.set({
    name: ACCESS_TOKEN_COOKIE,
    value: accessToken,
    httpOnly: true,
    secure: isSecureCookie(),
    sameSite: "lax",
    path: "/",
    maxAge: parseAccessTokenTTL(expiresAt),
  });

  response.cookies.set({
    name: REFRESH_TOKEN_COOKIE,
    value: refreshToken,
    httpOnly: true,
    secure: isSecureCookie(),
    sameSite: "lax",
    path: "/",
    maxAge: DEFAULT_REFRESH_TTL_SECONDS,
  });

  response.cookies.set({
    name: CSRF_TOKEN_COOKIE,
    value: csrfToken,
    httpOnly: false,
    secure: isSecureCookie(),
    sameSite: "lax",
    path: "/",
    maxAge: DEFAULT_REFRESH_TTL_SECONDS,
  });
}

export function clearAuthCookies(response: NextResponse): void {
  response.cookies.set({
    name: ACCESS_TOKEN_COOKIE,
    value: "",
    httpOnly: true,
    secure: isSecureCookie(),
    sameSite: "lax",
    path: "/",
    maxAge: 0,
  });

  response.cookies.set({
    name: REFRESH_TOKEN_COOKIE,
    value: "",
    httpOnly: true,
    secure: isSecureCookie(),
    sameSite: "lax",
    path: "/",
    maxAge: 0,
  });

  response.cookies.set({
    name: CSRF_TOKEN_COOKIE,
    value: "",
    httpOnly: false,
    secure: isSecureCookie(),
    sameSite: "lax",
    path: "/",
    maxAge: 0,
  });
}