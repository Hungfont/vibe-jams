---
description: "Use when implementing or updating frontend form flows (UI -> validate -> call server). Enforces Zod client+server validation, controlled API Route or Server Action submission, typed error handling, and post-submit revalidation patterns."
applyTo: "frontend/src/**/*.ts, frontend/src/**/*.tsx, frontend/src/**/*.js, frontend/src/**/*.jsx"
---
# Frontend Form Submit Flow Policy

Use this policy for mutation flows where users submit data from UI forms.

## Required Rules

- Model form flow explicitly as: input capture -> client validation -> submit -> server validation -> response mapping -> UI update.
- Validate in two layers with Zod:
  - Client-side Zod validation before submit for fast UX feedback.
  - Server-side Zod validation as the source of truth before persisting or triggering side effects.
- For submit channels, allow either:
  - App Router API routes (`/api/**`) for HTTP-style mutation flow.
  - Server Actions for controlled in-app mutation flow.
- Keep request/response contracts typed and stable with a standard error envelope: `success`, `error.code`, `error.message`, `error.fieldErrors[]`.
- Prevent duplicate submissions by managing pending state (`isSubmitting`) and disabling submit actions while a request is in flight.
- Map server validation errors back to form fields and keep generic error messaging for unexpected failures.
- After successful mutation, refresh dependent UI data intentionally:
  - If the screen uses SWR, revalidate with `mutate` on affected keys.
  - If data was loaded on the server, trigger the intended refresh path for the current route.

## Disallowed Without Explicit User Approval

- Relying only on client-side validation for business-critical checks.
- Direct browser calls to backend service domains that bypass frontend API/data-access flow.
- Using both API Route and Server Action for the same mutation path without a clear ownership boundary.
- Silent failure handling (for example: swallowing errors and showing success UI).
- Submitting forms without pending-state protection, which can cause double-submit behavior.
- Ad-hoc, inconsistent error response shapes across form endpoints.

## Implementation Notes

- Keep API handlers in App Router route files (`app/api/**/route.ts`) with explicit HTTP status codes.
- For Server Actions, keep mutation logic deterministic and return typed results that can be mapped to form-level and field-level UI states.
- Enforce one consistent error envelope that supports both field-level and form-level errors: `success`, `error.code`, `error.message`, `error.fieldErrors[]`.
- Keep Zod schemas close to the feature and reuse them where practical to reduce drift between UI and server checks.
- For authenticated submissions, centralize token/session forwarding inside frontend-owned server layers rather than in client components.