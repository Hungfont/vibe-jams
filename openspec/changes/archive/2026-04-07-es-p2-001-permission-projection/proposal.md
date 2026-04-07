## Why

Guest control permissions are currently all-or-nothing and effectively tied to host-only command policies, which prevents hosts from granting granular controls while maintaining deterministic authorization behavior. ES-P2-001 introduces permission projection so hosts can manage guest control capabilities in realtime, with auditable updates and command enforcement that reuses centralized authZ decisions from ES-P2-004.

## What Changes

- Add a session-scoped permission projection model for `canControlPlayback`, `canReorderQueue`, and `canChangeVolume`.
- Add host-managed permission update commands and persistence in jam-service runtime state.
- Publish permission update events to `jam.permission.events` using shared Kafka envelope contracts.
- Enforce guest command checks at command entrypoints by consuming centralized authZ decisions (no duplicate host/guest policy logic).
- Propagate permission updates into realtime fanout so active clients reflect toggles immediately.

## Capabilities

### New Capabilities
- `jam-permission-projection`: session-scoped guest permission state, host update commands, persistence, and event publication contract.

### Modified Capabilities
- `api-service-bff-microservice-routing`: delegated jam policy routes include permission command/update routes and enforce centralized decision checks before forwarding.
- `jam-playback-command-pipeline`: playback command authorization allows/denies guests based on projected permission state rather than host-only blanket policy.
- `jams-queue-command-handling`: queue reorder authorization allows/denies guests based on projected permission state while preserving optimistic concurrency guarantees.
- `kafka-event-foundation`: `jam.permission.events` topic and envelope compatibility requirements are added for permission projection updates.
- `realtime-fanout`: fanout consumer and websocket broadcast paths include permission update events for active room sync.

## Impact

- Backend jam-service: permission projection model, update handlers, storage persistence, and event publication.
- Backend api-service: delegated route and centralized authorization integration for permission update and protected command paths.
- Backend playback-service and queue command entry authorization paths: guest permission enforcement.
- Backend rt-gateway: consume and broadcast permission events.
- Frontend jam room controls: realtime reflection of permission toggles and immediate blocked-action feedback.