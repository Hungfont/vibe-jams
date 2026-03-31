## 1. Session Lifecycle Domain and Persistence

- [x] 1.1 Add jam session model types for lifecycle state, host ownership, membership, and permissions.
- [x] 1.2 Extend repository interfaces with create/join/leave/end operations and active-session validation helpers.
- [x] 1.3 Implement Redis key schema and atomic persistence for session metadata, members, and permissions.
- [x] 1.4 Add repository unit tests for lifecycle state transitions and concurrent membership updates.

## 2. HTTP and Service Lifecycle Endpoints

- [x] 2.1 Add HTTP routes/handlers for `POST /api/v1/jams/{jamId}/join` and `POST /api/v1/jams/{jamId}/leave`.
- [x] 2.2 Update create flow to persist host/session metadata and return stable payloads/status codes.
- [x] 2.3 Update end flow to enforce host ownership authorization before ending the session.
- [x] 2.4 Implement host-leave behavior that ends the session with explicit end cause.

## 3. Write Gating and Authorization Integration

- [x] 3.1 Add shared service-level guard to reject queue mutations when session state is ended.
- [x] 3.2 Add playback command guard to reject commands when session state is ended.
- [x] 3.3 Preserve existing auth entitlement contract (`401 unauthorized`, `403 premium_required`) for protected lifecycle operations.
- [x] 3.4 Standardize session-ended error mapping and payload shape for queue/playback write rejections.

## 4. Eventing, Contract Updates, and Verification

- [x] 4.1 Publish lifecycle transition events (`create`, `join`, `leave`, `end`) to `jam.session.events`.
- [x] 4.2 Update jam-service OpenAPI/service docs for new lifecycle endpoints and ownership/state constraints.
- [x] 4.3 Add handler/API integration tests for success and error matrix (auth failures, non-host end, host leave ends session).
- [x] 4.4 Add cross-path integration tests validating that writes are blocked after session end.
- [x] 4.5 Run `go test ./...` for `backend/jam-service` and affected playback/queue packages and fix regressions.
