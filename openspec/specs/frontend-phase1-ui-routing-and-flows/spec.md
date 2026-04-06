# frontend-phase1-ui-routing-and-flows Specification

## Purpose
TBD - created by archiving change phase1-spotify-like-jam-ui-pages. Update Purpose after archive.
## Requirements
### Requirement: Frontend SHALL define Phase 1 page URLs with explicit Jam UX ownership
The frontend SHALL provide URL-scoped page behavior for lobby and jam room experiences so each page has deterministic purpose, data dependencies, and action boundaries.

#### Scenario: Lobby URL ownership
- **WHEN** a user navigates to `/`
- **THEN** the frontend provides create and join Jam actions with deterministic validation and transition behavior

#### Scenario: Jam room URL ownership
- **WHEN** a user navigates to `/jam/{jamId}`
- **THEN** the frontend provides room shell behavior with queue, playback, participants, and diagnostics views bound to that session

### Requirement: Browser calls SHALL use frontend API routes as the only backend access boundary
The frontend SHALL route all browser-initiated requests through App Router API endpoints and MUST NOT call backend service domains directly from browser components.

#### Scenario: Browser submits protected action
- **WHEN** a user triggers create, join, queue, playback, or orchestration actions
- **THEN** the browser calls a frontend API route that forwards session context and maps responses to normalized envelope fields

#### Scenario: Direct backend URL usage attempted
- **WHEN** a browser component attempts to call backend service endpoints directly
- **THEN** the implementation is considered non-compliant with this requirement

### Requirement: Lobby page SHALL support create and join flows with auth and entitlement gating
The lobby page SHALL provide create and join actions, SHALL validate auth context before protected requests, and SHALL navigate to jam room on success.

#### Scenario: Create succeeds
- **WHEN** an entitled user submits create from `/`
- **THEN** frontend route flow returns `jamId` and the UI navigates to `/jam/{jamId}`

#### Scenario: Join validation fails
- **WHEN** a user submits join with invalid session input or unauthorized context
- **THEN** the lobby shows deterministic error feedback and remains on `/`

### Requirement: Jam room first render SHALL hydrate from orchestration and preserve core state under partial dependency issues
The jam room page SHALL use orchestration for first load and SHALL keep required session and queue sections visible even when optional dependency segments are degraded.

#### Scenario: Orchestration full success
- **WHEN** orchestration response is successful without degradation
- **THEN** the room renders session, participants, queue, and playback context from one payload

#### Scenario: Orchestration partial result
- **WHEN** orchestration returns `partial=true` with dependency issues
- **THEN** the room renders required core sections and shows actionable degraded notices for affected optional segments

### Requirement: Queue UI SHALL enforce version-aware mutation and deterministic conflict recovery
Queue add, remove, and reorder interactions SHALL execute through frontend routes, SHALL honor queue version behavior, and SHALL provide conflict recovery with snapshot refresh path.

#### Scenario: Queue reorder conflict
- **WHEN** reorder response indicates `version_conflict`
- **THEN** the UI refetches authoritative snapshot and offers retry without corrupting local projection

#### Scenario: Track validation error
- **WHEN** add flow encounters `track_not_found` or `track_unavailable`
- **THEN** the UI preserves existing queue state and surfaces explicit, actionable error feedback

### Requirement: Playback UI SHALL enforce host-only policy and deterministic command outcome handling
Playback controls SHALL only permit host command execution and SHALL map unauthorized, host-only, version conflict, and session-ended outcomes to deterministic UI states.

#### Scenario: Host command accepted
- **WHEN** host submits a valid playback command with required fields
- **THEN** UI shows accepted pending state and waits for realtime authoritative transition

#### Scenario: Non-host command attempt
- **WHEN** non-host user attempts playback command
- **THEN** command is blocked with host-only reason and no state transition is applied

### Requirement: Realtime synchronization SHALL apply monotonic event updates and snapshot fallback recovery
Jam room realtime state SHALL process only monotonic aggregateVersion updates, SHALL ignore stale or duplicate events, and SHALL recover via snapshot when gaps or stale reconnect cursors are detected.

#### Scenario: Monotonic event stream
- **WHEN** incoming event version equals local version plus one
- **THEN** the event is applied to room projection and local version advances

#### Scenario: Gap or stale reconnect
- **WHEN** gap is detected or reconnect cursor is stale
- **THEN** the frontend fetches authoritative session snapshot, replaces local projection, and resumes incremental processing

### Requirement: Error and degraded diagnostics UI SHALL preserve room usability with deterministic mappings
The diagnostics surface SHALL map normalized error codes to user actions while preserving room usability when required data remains valid.

#### Scenario: Blocking auth state
- **WHEN** room receives `unauthorized` state
- **THEN** UI prompts re-auth flow and blocks protected writes

#### Scenario: Dependency degradation
- **WHEN** dependency timeout or unavailable status affects optional segments
- **THEN** UI shows degraded panel and keeps required room capabilities available

### Requirement: Phase 1 UI SHALL use shadcn components to provide Spotify-like interaction structure
The frontend SHALL compose lobby and room surfaces using approved shadcn primitives and Tailwind utilities to achieve Spotify-like information hierarchy and interaction behavior, and MUST enforce this policy through frontend validation checks for new code.

#### Scenario: Lobby composition
- **WHEN** rendering lobby create/join page
- **THEN** UI uses approved shadcn Card, Tabs, Input, Button, Alert, and Toast primitives for action and feedback

#### Scenario: Jam room composition
- **WHEN** rendering jam room and focused views
- **THEN** UI uses approved shadcn Card, Tabs, Badge, ScrollArea, Slider, Tooltip, Dialog, Skeleton, and Alert patterns for queue, playback, participants, and diagnostics

#### Scenario: New duplicate primitive introduced
- **WHEN** a change introduces a custom primitive duplicating an approved shadcn-equivalent primitive category
- **THEN** frontend conformance validation fails and requires primitive reuse or approved exception

### Requirement: OpenSpec workflow execution SHALL bind to custom agent context with mapped resources
OpenSpec propose/apply/explore/archive flows for this capability SHALL resolve execution context (Frontend Engineer, Backend Engineer, or Agent) and SHALL load corresponding rules, instructions, skills, and prompts before artifact generation or code implementation.

#### Scenario: Frontend scope context resolution
- **WHEN** workflow scope targets `frontend/**`
- **THEN** execution uses Frontend Engineer context and loads `.github/rules/frontend/**`, `.github/instructions/frontend/fe-*.instructions.md`, and relevant frontend skills/prompts

#### Scenario: Backend scope context resolution
- **WHEN** workflow scope targets `backend/**`
- **THEN** execution uses Backend Engineer context and loads `.github/rules/backend/**`, `.github/instructions/backend/be-*.instructions.md`, and relevant backend skills

#### Scenario: OpenSpec orchestration context resolution
- **WHEN** workflow scope is OpenSpec artifact orchestration or cross-domain coordination
- **THEN** execution uses Agent context and loads `.github/rules/common/**` plus OpenSpec workflow skills

### Requirement: Frontend SHALL expose auth session routes through frontend-owned API boundary
The frontend SHALL expose auth lifecycle API routes and SHALL keep browser-to-backend communication within frontend-owned `/api/**` routes.

#### Scenario: Browser login submission
- **WHEN** a user submits login form from frontend login page
- **THEN** browser calls `POST /api/auth/login`, and frontend route forwards request to auth-service public login endpoint

#### Scenario: Browser session lifecycle actions
- **WHEN** frontend triggers refresh, logout, or me retrieval
- **THEN** browser calls `POST /api/auth/refresh`, `POST /api/auth/logout`, or `GET /api/auth/me` and routes normalize success/error envelopes

### Requirement: Login page SHALL provide Spotify-like visual hierarchy with accessible form behavior
The frontend SHALL render `/login` using Spotify-like dark visual hierarchy and SHALL preserve accessibility and deterministic submit state behavior.

#### Scenario: Login page first render
- **WHEN** user navigates to `/login`
- **THEN** page renders brand-forward dark layout, credential form, and deterministic loading/error/success UI states

#### Scenario: Login form validation and submit
- **WHEN** user submits invalid or valid credentials
- **THEN** client-side zod validation blocks invalid payloads, valid payloads submit through frontend auth route, and server errors map to actionable messages

