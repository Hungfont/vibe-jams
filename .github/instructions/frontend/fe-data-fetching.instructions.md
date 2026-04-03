---
description: "Use when implementing or updating frontend data fetching in Next.js App Router. Enforces Server Component-first loading, SWR-only client revalidation, shared fetch utilities, and intentional caching/error handling."
applyTo:
  - "frontend/src/**/*.ts"
  - "frontend/src/**/*.tsx"
  - "frontend/src/**/*.js"
  - "frontend/src/**/*.jsx"
---
# Frontend Data Fetching Policy

Use predictable and explicit data-loading patterns for Next.js App Router code.

## Required Rules

- Prefer Server Component data fetching for initial page data. Add "use client" only when interactivity requires client-side hooks.
- For client-side reads (polling, focus revalidation, or user-driven refresh), use SWR. Treat SWR as mandatory unless the user explicitly approves an exception.
- In Client Components, do not implement initial/read data loading with `useEffect + fetch` or ad-hoc promise chains. Use `useSWR(key, fetcher, options)`.
- Reuse shared fetch utilities (for example: frontend/src/lib/fetcher.ts) instead of duplicating low-level fetch handling in many components.
- Keep SWR keys deterministic (stable string/tuple keys). Avoid non-deterministic keys that break caching and revalidation behavior.
- Handle HTTP failures explicitly (`response.ok` checks) and surface actionable UI states (`loading`, `error`, `success`).
- Keep API response shapes stable and typed. Use consistent envelope patterns when calling internal APIs.
- Framework caching defaults are acceptable for low-risk paths, but set explicit caching (`cache`, `next.revalidate`, or tags) when stale data can affect UX or correctness.
- For frontend-to-backend flow, keep browser calls inside the frontend project boundary by using frontend API routes (`app/api/**/route.ts`) or frontend-owned clients that centralize URL/auth handling.

## Disallowed Without Explicit User Approval

- Introducing a second client data-fetching library when SWR already satisfies the use case.
- Client-side read flows implemented with raw `fetch` in `useEffect` instead of SWR.
- Scattering ad-hoc `fetch` logic across many UI components when shared fetch helpers can be used.
- Silent error swallowing (for example: returning empty data without surfacing request failure).
- Mixing conflicting caching strategies in the same feature without documenting intent.
- Direct browser calls to backend services that bypass the frontend project's API/data-access flow without explicit approval.

## Implementation Notes

- Keep endpoint handlers in App Router route files (`app/api/**/route.ts`) with clear status codes and typed payloads.
- For internal browser requests, prefer relative API paths (for example: `/api/status`).
- When polling is enabled in SWR, set an explicit interval and disable focus revalidation only when the feature requires it.
- For non-read operations (`POST`, `PUT`, `PATCH`, `DELETE`), use API Route or Server Action submit flows, then revalidate with `mutate(...)` instead of treating SWR as a mutation transport layer.