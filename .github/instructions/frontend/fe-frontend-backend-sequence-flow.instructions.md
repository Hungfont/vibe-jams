---
description: "Use when implementing or updating frontend jam flow, API routes, orchestration, realtime websocket bootstrap, or backend service calls. Enforces exact alignment with docs/frontend-backend-sequence.md."
name: "Frontend Backend Sequence Flow Guardrail"
applyTo:
  - "frontend/src/app/api/**/*.ts"
  - "frontend/src/app/jam/**/*.tsx"
  - "frontend/src/components/jam/**/*.tsx"
  - "frontend/src/lib/api/**/*.ts"
  - "frontend/src/lib/jam/**/*.ts"
---
# Frontend-Backend Sequence Flow Guardrail

The flow in docs/frontend-backend-sequence.md is the source of truth.
ALWAYS All frontend flows that call backend services (including API routes, orchestration logic, realtime bootstrap, and service calls) must strictly follow the sequence and contracts defined in this document. Any deviation from this flow requires explicit approval.
## Required Rules

- Keep frontend request flow aligned with docs/frontend-backend-sequence.md exactly.
- Keep browser-side calls on frontend-owned endpoints (`/api/**`). Do not call backend service base URLs directly from client components.
- Keep route-to-service mapping consistent with the table in docs/frontend-backend-sequence.md.
- Preserve BFF orchestration flow:
  - Frontend route: `POST /api/bff/jam/{jamId}/orchestration`
  - Upstream BFF route: `POST /v1/bff/mvp/sessions/{jamId}/orchestration`
  - Required dependencies: auth-service and jam-service
  - Optional dependencies: catalog-service and playback-service
- Preserve realtime bootstrap flow:
  - Get websocket config via `GET /api/realtime/ws-config`
  - Connect websocket directly to rt-gateway `/ws` with `sessionId` and `lastSeenVersion`
- For protected jam mutations, keep auth validation in frontend server-side flow before upstream mutation calls.

## Disallowed Without Explicit User Approval

- Changing existing frontend endpoint paths or upstream backend paths defined in docs/frontend-backend-sequence.md.
- Rewiring a mapped route to a different backend service than documented.
- Bypassing frontend API routes from browser code.
- Changing BFF dependency behavior (required vs optional) without explicit request.

## Change Management Rule

- If a new flow is introduced or an existing flow changes, update docs/frontend-backend-sequence.md in the same change.
- The docs update must include:
  - sequence diagram update (HTTP or realtime as applicable)
  - frontend route to upstream mapping update
  - short note describing behavior change
