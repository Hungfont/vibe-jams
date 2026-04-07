# Frontend -> Backend Service Call Sequence

This document reflects the current implementation flow for frontend requests and realtime updates.

## How To Use With Guardrail

- Treat .github/instructions/frontend/fe-frontend-backend-sequence-flow.instructions.md as the implementation guardrail.
- Treat this document as the executable contract for frontend page flows and route-to-service mapping.
- For any frontend jam-flow change, update code and this doc in one change so behavior and documentation stay in sync.

## Page-Ordered Flow (Runtime Order)

The sections below are intentionally ordered by page runtime path.

### 1) Login Page (`/login`)

- Entry file: frontend/src/app/login/page.tsx
- Client feature: frontend/src/components/auth/login-form.tsx
- Primary user actions:
    - Submit identity/password login form
    - Receive normalized error feedback for invalid credentials
- Frontend API routes used in order:
    - `POST /api/auth/login`
    - `POST /api/auth/refresh` (session renewal path)
    - `POST /api/auth/logout` (session termination path)
    - `GET /api/auth/me` (current-claims lookup)
- UI primitives used from frontend/src/components/ui:
    - Card, Input, Button, Alert, Separator, Toast

### 2) Lobby Page (`/`)

- Entry file: frontend/src/app/page.tsx
- Client feature: frontend/src/components/jam/lobby-client.tsx
- Primary user actions:
    - Create Jam
    - Join Jam
- Frontend API routes used in order:
    - `POST /api/auth/validate` (create pre-check)
    - `POST /api/jam/create` (create flow)
    - `POST /api/jam/{jamId}/join` (join flow)
- UI primitives used from frontend/src/components/ui:
    - Card, Tabs, Input, Button, Alert, Toast

### 3) Jam Page SSR Bootstrap (`/jam/{jamId}`)

- Entry file: frontend/src/app/jam/[jamId]/page.tsx
- Server-side bootstrap action:
    - `POST /api/bff/jam/{jamId}/orchestration`
- Purpose:
    - Load initial room state before client hydration
    - Pass initial view and initial data/error to JamRoomClient

### 4) Jam Room Client Runtime

- Client feature: frontend/src/components/jam/jam-room-client.tsx
- Runtime flow order:
    - Hydrate with SSR orchestration data
    - Start periodic orchestration refresh (SWR)
    - Bootstrap websocket config via `GET /api/realtime/ws-config`
    - Open websocket to rt-gateway `/ws`
    - Process realtime events and run snapshot recovery when needed
- User action flows:
    - Queue actions: add/remove/reorder
    - Moderation actions (host): mute/kick participants
    - Permission projection updates (host): playback/reorder/volume guest toggles
    - Playback commands: play/pause/next/prev/seek
    - End session (host)
- Frontend API routes used:
    - `GET /api/jam/{jamId}/state`
    - `GET /api/jam/{jamId}/queue/snapshot`
    - `POST /api/jam/{jamId}/queue/add`
    - `POST /api/jam/{jamId}/queue/remove`
    - `POST /api/jam/{jamId}/queue/reorder`
    - `POST /api/jam/{jamId}/moderation/mute`
    - `POST /api/jam/{jamId}/moderation/kick`
    - `GET /api/jam/{jamId}/permissions`
    - `POST /api/jam/{jamId}/permissions`
    - `POST /api/jam/{jamId}/playback/commands`
    - `POST /api/jam/{jamId}/end`
- UI primitives used from frontend/src/components/ui:
    - Layout/status: Badge, Separator, Skeleton
    - Inputs/actions: Input, Button, Slider
    - Navigation/content: Tabs, Card, ScrollArea
    - Interactive overlays: Dialog, DropdownMenu, Tooltip
    - Feedback/identity: Alert, Toast, Avatar

## UI Component Usage Approach (By Page)

When implementing or reviewing jam flows, keep UI usage tied to page responsibility and flow order.

| Page | UX responsibility | UI primitives (frontend/src/components/ui) |
| --- | --- | --- |
| `/login` | Authentication entry and credential session bootstrap | Card, Input, Button, Alert, Separator, Toast |
| `/` Lobby | Session entry and entitlement-safe create/join | Card, Tabs, Input, Button, Alert, Toast |
| `/jam/{jamId}` SSR | Initial orchestration bootstrap and error handoff | Alert (error fallback) |
| `/jam/{jamId}` Room | Realtime collaboration and host controls | Tabs, Card, ScrollArea, DropdownMenu, Tooltip, Slider, Dialog, Badge, Avatar, Alert, Toast, Button, Input, Skeleton, Separator |

## Guardrail Compliance Checklist

- Page order preserved: Lobby -> Jam SSR bootstrap -> Jam room runtime.
- Page order preserved: Login -> Lobby -> Jam SSR bootstrap -> Jam room runtime.
- Browser requests stay on frontend-owned `/api/**` routes.
- Route-to-service mapping matches this document.
- BFF orchestration semantics preserved:
    - Required: auth-service + jam-service
    - Optional/degradable: catalog-service + playback-service
- Realtime bootstrap preserved:
    - `GET /api/realtime/ws-config`
    - websocket connect to rt-gateway `/ws`
- Any new page flow or endpoint mapping change updates this document in the same PR.

## Scope

- Frontend client and Next.js App Router route handlers under frontend/src/app/api/**
- Backend services under backend/**:
  - auth-service
  - jam-service
  - playback-service
  - catalog-service
  - api-service (BFF)
  - rt-gateway
- Kafka fanout path used by realtime updates

## Service Base URLs (frontend config defaults)

- api-gateway (public ingress): http://localhost:8085
- auth-service: http://localhost:8081
- jam-service: http://localhost:8080
- playback-service: http://localhost:8082
- catalog-service: http://localhost:8083
- api-service (BFF, internal only): http://localhost:8084
- rt-gateway: http://localhost:8086

> **Note**: `api-gateway` (port 8085) is the sole public entry point for all BFF and auth flows. `api-service` (port 8084) is internal-only and not reachable from browsers or frontend servers directly.

## Sequence 1: HTTP Request/Response Flow

```mermaid
sequenceDiagram
    autonumber
    actor U as Browser (Frontend UI)
    participant N as Next.js API Routes (frontend/src/app/api)
    participant G as api-gateway
    participant A as auth-service
    participant J as jam-service
    participant P as playback-service
    participant C as catalog-service
    participant B as api-service (BFF)

    Note over U,N: 0) AuthN lifecycle on frontend auth boundary
    U->>N: POST /api/auth/login
    N->>G: POST /v1/auth/login
    G->>A: POST /v1/auth/login
    A-->>G: access+refresh token pair, claims
    G-->>N: upstream response
    N-->>U: Envelope success/error + HttpOnly cookie set

    U->>N: POST /api/auth/refresh (X-CSRF-Token)
    N->>G: POST /v1/auth/refresh
    G->>A: POST /v1/auth/refresh
    A-->>G: rotated token pair, claims
    G-->>N: upstream response
    N-->>U: Envelope success/error + rotated cookies

    U->>N: POST /api/auth/logout (X-CSRF-Token)
    N->>G: POST /v1/auth/logout
    G->>A: POST /v1/auth/logout
    A-->>G: status
    G-->>N: upstream response
    N-->>U: Envelope success/error + cookie clear

    U->>N: GET /api/auth/me
    N->>G: GET /v1/auth/me
    G->>A: GET /v1/auth/me
    A-->>G: normalized claims
    G-->>N: upstream response
    N-->>U: Envelope success/error

    Note over U,N: 1) Auth validation API route (authN/authZ flow)
    U->>N: POST /api/auth/validate
    N->>G: POST /internal/v1/auth/validate
    G->>A: POST /internal/v1/auth/validate
    A-->>G: 200 claims or 401
    G-->>N: upstream response
    N-->>U: Envelope success/error

    Note over U,N: 2) Mandatory BFF-first hop for all microservice HTTP calls
    U->>N: POST /api/jam/create (or join/leave/end/queue/*/moderation/*)
    N->>G: POST /v1/bff/mvp/sessions/{jamId}/commands
    G->>A: POST /internal/v1/auth/validate (token verification)
    A-->>G: 200 claims or 401
    G->>B: POST /v1/bff/mvp/sessions/{jamId}/commands (X-Auth-* headers)
    B->>J: /api/v1/jams/... endpoint
    J-->>B: session/queue response
    B-->>G: normalized result
    G-->>N: upstream response
    N-->>U: Envelope success/error

    U->>N: POST /api/jam/{jamId}/playback/commands
    N->>G: POST /v1/bff/mvp/sessions/{jamId}/playback/commands
    G->>B: POST /v1/bff/mvp/sessions/{jamId}/playback/commands (X-Auth-* headers)
    B->>P: POST /v1/jam/sessions/{jamId}/playback/commands
    P-->>B: 202 accepted or error
    B-->>G: normalized result
    G-->>N: upstream response
    N-->>U: Envelope success/error

    U->>N: GET /api/jam/{jamId}/permissions
    N->>G: GET /api/v1/jams/{jamId}/permissions
    G->>B: GET /api/v1/jams/{jamId}/permissions (X-Auth-* headers)
    B->>J: GET /api/v1/jams/{jamId}/permissions
    J-->>B: SessionPermissions
    B-->>G: normalized result
    G-->>N: upstream response
    N-->>U: Envelope success/error

    U->>N: POST /api/jam/{jamId}/permissions
    N->>G: POST /api/v1/jams/{jamId}/permissions
    G->>B: POST /api/v1/jams/{jamId}/permissions (X-Auth-* headers)
    B->>J: POST /api/v1/jams/{jamId}/permissions
    J-->>B: updated SessionPermissions or deterministic 403/400
    B-->>G: normalized result
    G-->>N: upstream response
    N-->>U: Envelope success/error

    U->>N: GET /api/catalog/tracks/{trackId}
    N->>G: GET /v1/bff/mvp/catalog/tracks/{trackId}
    G->>B: GET /v1/bff/mvp/catalog/tracks/{trackId}
    B->>C: GET /internal/v1/catalog/tracks/{trackId}
    C-->>B: lookup response or 404
    B-->>G: normalized result
    G-->>N: upstream response
    N-->>U: Envelope success/error

    Note over U,N: 5) BFF orchestration (SSR + room bootstrap) — flows through api-gateway
    U->>N: POST /api/bff/jam/{jamId}/orchestration
    N->>G: POST /v1/bff/mvp/sessions/{jamId}/orchestration (Bearer token)
    G->>A: POST /internal/v1/auth/validate (token verification)
    A-->>G: 200 claims or 401
    G->>B: POST /v1/bff/mvp/sessions/{jamId}/orchestration (X-Auth-* headers, Authorization preserved for compatibility)
    B->>J: GET /api/v1/jams/{jamId}/state (X-Auth-* forwarded, required)
    opt trackId present
        B->>C: GET /internal/v1/catalog/tracks/{trackId} (optional)
        C-->>B: track lookup or degradation issue
    end
    B-->>G: OrchestrateData (partial possible)
    G-->>N: upstream response
    N-->>U: Envelope success/error
```

## Sequence 2: Realtime WebSocket + Kafka Fanout Flow

```mermaid
sequenceDiagram
    autonumber
    actor U as Browser (Jam Room)
    participant N as Next.js API Routes
    participant G as api-gateway
    participant B as api-service (BFF)
    participant R as rt-gateway
    participant K as Kafka
    participant J as jam-service
    participant P as playback-service

    Note over U,N: 1) Realtime bootstrap config via mandatory BFF-first HTTP hop
    U->>N: GET /api/realtime/ws-config?sessionId=...&lastSeenVersion=...
    N->>G: GET /v1/bff/mvp/realtime/ws-config?sessionId=...&lastSeenVersion=... (Authorization or auth cookies forwarded)
    G->>B: GET /v1/bff/mvp/realtime/ws-config?sessionId=...&lastSeenVersion=...
    B->>R: GET /internal/v1/realtime/ws-config?sessionId=...&lastSeenVersion=...
    R-->>B: { wsUrl, sessionId, lastSeenVersion }
    B-->>G: normalized result
    G-->>N: upstream response
    N-->>U: { wsUrl, sessionId, lastSeenVersion }

    Note over U,R: 2) Client opens websocket through gateway/BFF proxy path
    U->>G: WS /v1/bff/mvp/realtime/ws?sessionId=...&lastSeenVersion=...
    G->>B: WS /v1/bff/mvp/realtime/ws?sessionId=...&lastSeenVersion=...
    B->>R: WS /ws?sessionId=...&lastSeenVersion=...

    par Event production
        J->>K: Publish jam.session.* and jam.queue.* events
    and
        J->>K: Publish jam.moderation.* events
    and
        J->>K: Publish jam.permission.events
    and
        P->>K: Publish jam.playback.updated events
    end

    R->>K: Consume queue/playback/moderation/permission topics
    R-->>U: Fanout outbound realtime events

    alt Gap or stale cursor detected
        R->>J: GET /api/v1/jams/{sessionId}/state (snapshot recovery)
        J-->>R: SessionStateSnapshot
        R-->>U: jam.session.snapshot (recovery=true)
    end
```

## Frontend Route -> Backend Endpoint Mapping

| Frontend route | Backend service | Upstream endpoint |
| --- | --- | --- |
| POST /api/auth/login | api-gateway -> auth-service | POST /v1/auth/login |
| POST /api/auth/refresh | api-gateway -> auth-service | POST /v1/auth/refresh |
| POST /api/auth/logout | api-gateway -> auth-service | POST /v1/auth/logout |
| GET /api/auth/me | api-gateway -> auth-service | GET /v1/auth/me |
| POST /api/auth/validate | api-gateway -> auth-service | POST /internal/v1/auth/validate |
| POST /api/jam/create | api-gateway -> api-service (BFF) -> jam-service | POST /api/v1/jams/create |
| POST /api/jam/{jamId}/join | api-gateway -> api-service (BFF) -> jam-service | POST /api/v1/jams/{jamId}/join |
| POST /api/jam/{jamId}/leave | api-gateway -> api-service (BFF) -> jam-service | POST /api/v1/jams/{jamId}/leave |
| POST /api/jam/{jamId}/end | api-gateway -> api-service (BFF) -> jam-service | POST /api/v1/jams/{jamId}/end |
| GET /api/jam/{jamId}/state | api-gateway -> api-service (BFF) -> jam-service | GET /api/v1/jams/{jamId}/state |
| GET /api/jam/{jamId}/queue/snapshot | api-gateway -> api-service (BFF) -> jam-service | GET /api/v1/jams/{jamId}/queue/snapshot |
| POST /api/jam/{jamId}/queue/add | api-gateway -> api-service (BFF) -> jam-service | POST /api/v1/jams/{jamId}/queue/add |
| POST /api/jam/{jamId}/queue/remove | api-gateway -> api-service (BFF) -> jam-service | POST /api/v1/jams/{jamId}/queue/remove |
| POST /api/jam/{jamId}/queue/reorder | api-gateway -> api-service (BFF) -> jam-service | POST /api/v1/jams/{jamId}/queue/reorder |
| POST /api/jam/{jamId}/moderation/mute | api-gateway -> api-service (BFF) -> jam-service | POST /api/v1/jams/{jamId}/moderation/mute |
| POST /api/jam/{jamId}/moderation/kick | api-gateway -> api-service (BFF) -> jam-service | POST /api/v1/jams/{jamId}/moderation/kick |
| GET /api/jam/{jamId}/permissions | api-gateway -> api-service (BFF) -> jam-service | GET /api/v1/jams/{jamId}/permissions |
| POST /api/jam/{jamId}/permissions | api-gateway -> api-service (BFF) -> jam-service | POST /api/v1/jams/{jamId}/permissions |
| POST /api/jam/{jamId}/playback/commands | api-gateway -> api-service (BFF) -> playback-service | POST /v1/jam/sessions/{jamId}/playback/commands |
| GET /api/catalog/tracks/{trackId} | api-gateway -> api-service (BFF) -> catalog-service | GET /internal/v1/catalog/tracks/{trackId} |
| POST /api/bff/jam/{jamId}/orchestration | api-gateway -> api-service (BFF) | POST /v1/bff/mvp/sessions/{jamId}/orchestration |
| GET /api/realtime/ws-config | api-gateway -> api-service (BFF) | GET /v1/bff/mvp/realtime/ws-config |
| Browser websocket connect | api-gateway -> api-service (BFF) -> rt-gateway | WS /v1/bff/mvp/realtime/ws -> /ws |

## Notes

- Frontend route handlers normalize backend errors into a common envelope.
- Frontend auth routes issue `auth_token` and `refresh_token` as HttpOnly cookies and enforce CSRF header matching for refresh/logout.
- auth-service token validation remains the shared gate for protected mutations through api-gateway validation middleware.
- **api-gateway is the sole public ingress**. It validates Bearer tokens via `POST /internal/v1/auth/validate` on auth-service, preferring `Authorization` and falling back to `auth_token`/`token` cookies when header auth is absent, then injects `X-Auth-UserId`, `X-Auth-Plan`, `X-Auth-SessionState`, `X-Auth-Scope` headers before forwarding to api-service.
- api-service reads identity from `X-Auth-*` headers only. It does not call auth-service. Requests without `X-Auth-UserId` are rejected with `401 unauthorized`.
- Non-auth frontend microservice HTTP calls are routed via api-service (BFF) before downstream delegation.
- Frontend websocket connect path is routed through `api-gateway -> api-service (BFF)` via `WS /v1/bff/mvp/realtime/ws`; direct browser websocket calls to rt-gateway `/ws` are non-compliant.
- During migration compatibility, api-gateway preserves `Authorization` (including cookie-derived bearer fallback) while also injecting `X-Auth-*` headers.
- api-service BFF treats jam as a required dependency; catalog can degrade and still return partial orchestration data.
- Orchestration is side-effect free. Playback mutations are accepted only on `POST /api/jam/{jamId}/playback/commands`; sending `playbackCommand` to orchestration returns `400 invalid_input`.
- rt-gateway fanout uses Kafka events and can recover client gaps by fetching jam-service state snapshots.

## Operational API Docs Endpoints

- api-gateway Swagger UI: `GET /swagger`
- api-gateway OpenAPI JSON: `GET /swagger/openapi.json`
- api-service Swagger UI: `GET /swagger`
- api-service OpenAPI JSON: `GET /swagger/openapi.json`
