## Context

ES-P2-001 introduces granular guest control permissions for active jam sessions: `canControlPlayback`, `canReorderQueue`, and `canChangeVolume`. Current behavior is still effectively host-only for these command families, and ES-P2-004 established centralized authorization decision flow that this change must consume rather than duplicate.

This is a cross-cutting backend and integration change touching jam-service projection state, api-service delegated command enforcement, playback/queue command authorization paths, Kafka topic contracts, and realtime fanout. Active clients must observe permission toggles immediately.

## Goals / Non-Goals

**Goals:**
- Add a durable session-scoped permission projection model for guest capabilities.
- Allow host users to update guest permissions through deterministic command endpoints.
- Publish permission updates to `jam.permission.events` with shared envelope compatibility.
- Enforce playback/reorder/volume command authorization using centralized authZ decisions from ES-P2-004 plus projected permission state.
- Broadcast permission updates through realtime fanout so active sessions reflect toggles immediately.

**Non-Goals:**
- Replacing ES-P2-004 centralized authZ decision ownership.
- Redesigning moderation state model or command contracts unrelated to permissions.
- Introducing per-command custom policy engines outside existing guard flow.
- Implementing unrelated frontend redesign beyond permission state rendering/handling.

## Decisions

1. Centralize permission decision ownership at delegated command entrypoints
- Decision: evaluate host/guest role from centralized decision path, then evaluate guest capability from session permission projection for protected command families.
- Rationale: preserves ES-P2-004 boundary and prevents duplicate role logic.
- Alternatives considered:
  - Re-implement host/guest role checks in each service: rejected due to policy drift risk.
  - Move all checks to jam-service only: rejected because upstream fail-fast behavior is required.

2. Store permissions as session projection in jam-service state
- Decision: persist default and updated permission flags in jam-service durable session state and expose snapshot/read model for dependent services.
- Rationale: permission source of truth remains session-scoped and co-located with participant lifecycle.
- Alternatives considered:
  - Store permissions only in api-service memory: rejected due to durability and multi-instance inconsistency.

3. Event contract for permission updates
- Decision: publish permission changes to `jam.permission.events` keyed by `sessionId` with actor, target, capability set, and resulting version metadata.
- Rationale: enables realtime propagation and auditability.
- Alternatives considered:
  - Reuse moderation topic for permission events: rejected due to semantic mixing and consumer complexity.

4. Realtime sync via rt-gateway consumer extension
- Decision: extend fanout consumer to process permission events as session-scoped updates using existing ordering and gap-recovery semantics.
- Rationale: active clients already depend on fanout ordering guarantees.
- Alternatives considered:
  - Poll permission state from clients: rejected due to delayed UX and unnecessary load.

## Risks / Trade-offs

- [Risk] Command authorization race between permission update and protected command execution -> Mitigation: include projection version metadata and rely on ordered event processing plus deterministic immediate server-side checks.
- [Risk] Additional authorization call/state lookup increases latency on delegated command paths -> Mitigation: reuse existing bounded dependency timeouts and short-circuit host paths.
- [Trade-off] New event topic increases operational footprint -> Mitigation: align provisioning and envelope contracts with existing Kafka foundation patterns.
- [Risk] Missing permission defaults could block legitimate guest actions unexpectedly -> Mitigation: define explicit default projection values at session creation and validate in tests.
