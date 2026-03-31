## Context

The MVP needs queue command handling for jam sessions with deterministic outcomes under concurrent client activity. Current scope requires add, remove, reorder, and snapshot operations with a strict `queueVersion` contract and idempotent add behavior. The service depends on Redis primitives and a shared API error model that already standardizes error payloads.

## Goals / Non-Goals

**Goals:**
- Provide a new `jams` microservice boundary for queue command ownership.
- Guarantee atomic queue mutation with monotonic `queueVersion` increments.
- Enforce idempotency for queue-add requests using client-provided idempotency keys.
- Reject stale reorder requests with `409 version_conflict`.
- Provide queue snapshots that always reflect the latest committed version.

**Non-Goals:**
- Implementing recommendation/ranking logic for tracks.
- Cross-session queue federation or multi-room orchestration.
- Long-term analytics and event replay infrastructure.

## Decisions

1. Service boundary: create a dedicated `jams` service
   - Decision: implement queue commands in a standalone microservice with its own API handlers and Redis repository.
   - Rationale: isolates queue consistency logic and keeps other services from duplicating write semantics.
   - Alternative considered: implement queue commands in `catalog-service`; rejected because catalog ownership is metadata-centric, not session queue mutation.

2. Atomic mutation strategy: Redis transaction/script for write + version bump
   - Decision: execute queue mutation and `queueVersion` increment in one atomic operation (Lua script or `MULTI/EXEC` guarded by keys).
   - Rationale: prevents partial updates and race conditions across concurrent command writers.
   - Alternative considered: application-side lock with separate Redis calls; rejected due to higher failure windows and operational complexity.

3. Version contract: optimistic concurrency for reorder
   - Decision: require `expectedQueueVersion` for reorder and return `409 version_conflict` when stale.
   - Rationale: explicit client retry semantics and deterministic ordering behavior.
   - Alternative considered: last-write-wins reorder; rejected because it hides conflicts and causes non-deterministic UX.

4. Idempotency: key-based dedup for queue-add
   - Decision: store idempotency key mapping per session command scope with short TTL and previous result reference.
   - Rationale: safely handles client/network retries without duplicate queue entries.
   - Alternative considered: dedup by track ID only; rejected because legitimate repeated tracks must remain possible.

5. Read model: snapshot includes queue and version in one response
   - Decision: expose queue snapshot endpoint returning ordered items plus current `queueVersion`.
   - Rationale: clients can synchronize local state and base next command on latest version.
   - Alternative considered: separate version endpoint; rejected due to extra round-trips and race potential.

## Risks / Trade-offs

- [High write contention on hot sessions] -> Mitigation: keep mutation scripts minimal, benchmark worst-case queue length, and set per-session request throttling if needed.
- [Idempotency key storage growth] -> Mitigation: TTL-based cleanup and bounded key namespace by session.
- [Cross-service error inconsistency] -> Mitigation: reuse shared error model and add contract tests for `409 version_conflict`.
- [Operational dependency on Redis availability] -> Mitigation: readiness checks, retry policy with bounded timeout, and clear failure responses.

## Migration Plan

1. Deploy `jams` service with read-only health endpoints.
2. Enable queue command endpoints behind a feature flag for internal traffic.
3. Route a small percentage of jam sessions to `jams` service and validate version/idempotency metrics.
4. Roll out fully once concurrency tests and production telemetry are stable.
5. Rollback: switch traffic back to previous queue path and disable feature flag; no irreversible data migration is required.

## Open Questions

- Should idempotency keys expire at a fixed global TTL or be configurable per client type?
- Do we require hard queue length limits in MVP, and if yes, what error code should represent overflow?
