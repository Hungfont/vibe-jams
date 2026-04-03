# Auth Service Requirement File

## Dependency rank
- Rank: 1
- Dependency type: foundational
- Depends on: none

## Backend APIs used
- POST /internal/v1/auth/validate
- GET /healthz

## Frontend requirements
1. Every protected frontend API route must forward cookie/session auth context to backend validation path.
2. Claims fields userId, plan, sessionState are mandatory before allowing protected UI actions.
3. Session state other than valid must map to unauthorized and trigger re-auth UX.
4. Premium plan claim must gate host create and host end actions before submission.

## Processing flow
1. User triggers protected action from UI.
2. Frontend API route extracts cookie/session and forwards auth context.
3. Auth validation response is mapped to normalized frontend envelope.
4. If unauthorized, frontend blocks action and renders re-auth prompt.
5. If valid, downstream service request continues.

## Spotify-like shadcn component mapping
- Alert: blocking unauthorized and premium messages.
- Toast: non-blocking auth warnings and retries.
- Badge: account tier and session state hint in room header.
- Tooltip: explain why an action is disabled.

## Frontend router and page mapping

### App pages
- / -> frontend/src/app/page.tsx
- /jam/[jamId] -> frontend/src/app/jam/[jamId]/page.tsx

### App API routes
- POST /api/auth/validate -> frontend/src/app/api/auth/validate/route.ts -> POST /internal/v1/auth/validate

### Router flow
1. User opens / or /jam/[jamId].
2. Protected action calls /api/auth/validate in server layer.
3. Route normalizes unauthorized and premium states before forwarding to downstream routes.

## Success criteria
1. Unauthorized behavior is deterministic across create, end, queue write, playback command.
2. Premium gating is deterministic for host-only premium actions.
3. No browser-direct backend call bypasses frontend API route guard.
