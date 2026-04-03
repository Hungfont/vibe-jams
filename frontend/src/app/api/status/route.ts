import { NextResponse } from "next/server";

interface StatusPayload {
  service: string;
  timestamp: string;
}

interface ApiResponse<T> {
  success: boolean;
  data: T;
}

export async function GET() {
  const payload: ApiResponse<StatusPayload> = {
    success: true,
    data: {
      service: "frontend",
      timestamp: new Date().toISOString(),
    },
  };

  return NextResponse.json(payload);
}
