import { NextResponse } from "next/server";
import { errorEnvelope, successEnvelope } from "@/lib/api/types";

export function jsonSuccess<T>(data: T, status = 200): NextResponse {
  return NextResponse.json(successEnvelope(data), { status });
}

export function jsonError(code: string, message: string, status: number, dependency?: string): NextResponse {
  return NextResponse.json(errorEnvelope(code, message, dependency), { status });
}
