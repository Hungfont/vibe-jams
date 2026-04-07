## MODIFIED Requirements

### Requirement: Frontend SHALL expose auth session routes through frontend-owned API boundary
The frontend SHALL expose auth lifecycle API routes and SHALL keep browser-to-backend communication within frontend-owned `/api/**` routes. Frontend auth routes (`POST /api/auth/login`, `POST /api/auth/refresh`, `POST /api/auth/logout`, `GET /api/auth/me`) MUST target api-gateway upstream endpoints rather than calling auth-service directly.

#### Scenario: Browser login submission through gateway-aligned frontend route
- **WHEN** a user submits login form from frontend login page
- **THEN** browser calls `POST /api/auth/login`, and frontend route forwards request to api-gateway path `POST /v1/auth/login`

#### Scenario: Browser session lifecycle actions through gateway-aligned frontend routes
- **WHEN** frontend triggers refresh, logout, or me retrieval
- **THEN** browser calls `POST /api/auth/refresh`, `POST /api/auth/logout`, or `GET /api/auth/me`
- **AND** frontend routes forward those requests to api-gateway paths `/v1/auth/refresh`, `/v1/auth/logout`, and `/v1/auth/me` with normalized envelope mapping

### Requirement: Jam room first render SHALL hydrate from orchestration and preserve core state under partial dependency issues
The jam room page SHALL use orchestration for first load and SHALL keep required session and queue sections visible even when optional dependency segments are degraded. Frontend orchestration route `POST /api/bff/jam/{jamId}/orchestration` MUST forward upstream to api-gateway path `POST /v1/bff/mvp/sessions/{jamId}/orchestration`.

#### Scenario: Orchestration request forwarded through api-gateway
- **WHEN** frontend server route receives `POST /api/bff/jam/{jamId}/orchestration`
- **THEN** route forwards request to api-gateway endpoint `POST /v1/bff/mvp/sessions/{jamId}/orchestration` and preserves normalized success/error envelope behavior

#### Scenario: Orchestration partial result remains compatible
- **WHEN** orchestration returns `partial=true` with dependency issues
- **THEN** the room renders required core sections and shows actionable degraded notices for affected optional segments
