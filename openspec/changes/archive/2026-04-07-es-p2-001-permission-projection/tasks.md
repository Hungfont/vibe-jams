## 1. Permission Projection Contract and Storage

- [x] 1.1 Add session permission projection model (`canControlPlayback`, `canReorderQueue`, `canChangeVolume`) to jam-service snapshot and durable repository state.
- [x] 1.2 Initialize deterministic permission defaults during session creation and preserve values across restart/reload paths.
- [x] 1.3 Add host-only permission update command interface and request validation in jam-service.

## 2. Permission Command APIs and Events

- [x] 2.1 Add jam-service permission update/read endpoints to expose and mutate permission projection.
- [x] 2.2 Publish accepted permission updates to Kafka topic `jam.permission.events` with shared envelope + actor metadata.
- [x] 2.3 Ensure rejected permission update attempts do not mutate state or publish permission events.

## 3. Authorization Enforcement on Protected Commands

- [x] 3.1 Integrate ES-P2-004 centralized authZ decision flow into permission update command handling without duplicating host/guest role logic.
- [x] 3.2 Enforce `canControlPlayback` for guest playback commands on delegated command entrypoint paths.
- [x] 3.3 Enforce `canReorderQueue` for guest queue reorder commands while preserving optimistic concurrency responses.
- [x] 3.4 Enforce `canChangeVolume` for guest volume command paths with deterministic forbidden permission errors.

## 4. API-Service and Realtime Integration

- [x] 4.1 Extend api-service delegated route coverage for permission projection update/read operations.
- [x] 4.2 Enforce permission-aware protected command forwarding behavior in api-service before downstream proxying.
- [x] 4.3 Extend rt-gateway fanout consumer to consume and broadcast `jam.permission.events` for active room subscribers.

## 5. Frontend Flow, Validation, and Documentation

- [x] 5.1 Update frontend jam API routes/client state to render and react to permission projection updates in realtime.
- [x] 5.2 Add deterministic client handling for permission-forbidden command responses.
- [x] 5.3 Update docs/runbooks/run.md with concise ES-P2-001 test flow and expected permission behavior.

## 6. Tests and Validation

- [x] 6.1 Add jam-service tests for permission projection defaults, updates, persistence, and event publication.
- [x] 6.2 Add api-service delegated-route tests for allowed/denied guest protected commands based on projection state.
- [x] 6.3 Add playback/queue command tests proving immediate block when guest permission is disabled.
- [x] 6.4 Add rt-gateway fanout test coverage for permission event broadcast.
- [x] 6.5 Execute targeted backend/frontend suites and record validation evidence in runbook update.