## Why

`ES-P1-001` defines MVP Jam session lifecycle APIs (`create`, `join`, `leave`, `end`), but the current implementation only partially covers lifecycle behavior and does not enforce host ownership/session-state constraints end-to-end. This change closes those gaps so session state is authoritative, consistent, and safe for downstream queue/playback writes.

## What Changes

- Add complete Jam lifecycle endpoints in `jam-service` for `create`, `join`, `leave`, and `end`.
- Enforce premium entitlement for host session creation and host ownership for explicit session end.
- Persist Jam session metadata and participant permissions in Redis keyspace with stable keys.
- Enforce host-leave semantics that transition session state to ended and reject subsequent write operations.
- Add unit and API/integration tests for success and error paths, including auth/ownership/state transitions.

## Capabilities

### New Capabilities
- `jam-session-lifecycle`: Session lifecycle APIs and state machine for create/join/leave/end, including ownership and write-gating after session end.

### Modified Capabilities
- `jams-queue-command-handling`: Queue command handling must reject writes when the related Jam session is ended.
- `jam-playback-command-pipeline`: Playback command handling must reject writes when the related Jam session is ended.

## Impact

- Affected code: `backend/jam-service` (handler/service/repository/model/router/tests), plus session-state checks in queue/playback command paths.
- API impact: New `join` and `leave` endpoints; tightened `end` authorization semantics; explicit state-related error outcomes.
- Dependencies: Existing auth claim contract from `es-p1-006-auth-entitlement-guard-jam`, Redis connectivity/key schema, and current Jam queue/playback components.
