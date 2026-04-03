"use client";

import useSWR from "swr";

import { fetcher } from "@/lib/fetcher";

interface StatusPayload {
  service: string;
  timestamp: string;
}

interface ApiResponse<T> {
  success: boolean;
  data: T;
}

export default function Home() {
  const { data, error, isLoading } = useSWR<ApiResponse<StatusPayload>>(
    "/api/status",
    fetcher,
    {
      refreshInterval: 10000,
      revalidateOnFocus: false,
    },
  );

  return (
    <main className="mx-auto flex min-h-screen w-full max-w-3xl flex-col justify-center px-6 py-12">
      <h1 className="text-3xl font-semibold tracking-tight">Frontend Bootstrap</h1>
      <p className="mt-2 text-sm text-slate-600">
        Next.js + Tailwind CSS + SWR is configured.
      </p>

      <section className="mt-8 rounded-lg border border-slate-200 bg-white p-5 shadow-sm">
        <h2 className="text-lg font-medium">SWR API Fetch Example</h2>

        {isLoading && (
          <p className="mt-3 text-sm text-slate-600">Loading /api/status ...</p>
        )}

        {error && (
          <p className="mt-3 text-sm text-red-600">Failed to load status.</p>
        )}

        {data && (
          <dl className="mt-4 grid gap-2 text-sm">
            <div>
              <dt className="font-medium">success</dt>
              <dd>{String(data.success)}</dd>
            </div>
            <div>
              <dt className="font-medium">service</dt>
              <dd>{data.data.service}</dd>
            </div>
            <div>
              <dt className="font-medium">timestamp</dt>
              <dd>{data.data.timestamp}</dd>
            </div>
          </dl>
        )}
      </section>
    </main>
  );
}
