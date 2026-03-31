## Context

`jam-service` currently supports queue and partial session endpoints, but session lifecycle behavior is incomplete for MVP: `join`/`leave` are missing, host ownership is not enforced for `end`, and there is no authoritative session state that blocks writes after host departure. Existing auth entitlement behavior from `es-p1-006-auth-entitlement-guard-jam` is available and should be reused rather than replaced.

This change affects multiple modules (`handler`, `service`, `repository`, and command-path checks) and introduces lifecycle state that other write paths depend on, so a shared design is required before implementation.

## Goals / Non-Goals

**Goals:**
- Implement complete lifecycle endpoints in `jam-service`: `create`, `join`, `leave`, and `end`.
- Persist session metadata and permissions in Redis-style keyspace with deterministic key naming.
- Enforce premium entitlement for `create` and host ownership for `end`.
- Ensure host leave or host end transitions session to `ended` and prevents further queue/playback writes.
- Provide deterministic API error/status behavior with unit and API/integration coverage.

**Non-Goals:**
- Replacing current auth-service integration or token issuance approach.
- Introducing distributed locking or multi-region consistency features beyond MVP.
- Redesigning queue/playback business logic beyond adding lifecycle-state gating.

## Decisions

1. **Introduce explicit Jam session aggregate state in repository layer**
   - Store session-level metadata separately from queue state:
     - `jam:{jamId}:session:meta` (status, hostUserId, createdAt, endedAt, endCause)
     - `jam:{jamId}:session:members` (participant membership set/map)
     - `jam:{jamId}:session:permissions` (role/permission map keyed by userId)
   - Why: lifecycle ownership/state checks should not be inferred from queue data.
   - Alternative considered: infer session state from queue mutations. Rejected due to weak semantics and inability to model membership.

2. **Apply lifecycle checks in service boundary for all writes**
   - `join`/`leave`/`end` and queue/playback write paths consult session state first.
   - Any write path targeting `ended` session returns deterministic conflict/invalid-state response.
   - Why: central guard keeps behavior consistent and prevents endpoint drift.
   - Alternative considered: enforce only at HTTP handlers. Rejected because internal command paths could bypass checks.

3. **Host ownership policy for explicit end; host leave implies end**
   - Explicit `end` requires authenticated actor to be the stored `hostUserId`.
   - `leave` by host transitions session to `ended` (`endCause=host_leave`).
   - `leave` by non-host removes membership and retains active session.
   - Why: matches ES-P1-001 DoD and makes termination semantics unambiguous.
   - Alternative considered: transfer host ownership on host leave. Rejected for MVP simplicity.

4. **Keep auth entitlement integration as-is, add ownership check after auth**
   - Reuse existing `401 unauthorized` and `403 premium_required` mapping for protected endpoints.
   - Perform ownership authorization after token/claim validation using stored session metadata.
   - Why: preserves completed `es-p1-006` contract and reduces auth regression risk.
   - Alternative considered: custom endpoint-specific auth mapping. Rejected due to contract divergence risk.

5. **Event semantics for lifecycle transitions**
   - Emit session events for create/join/leave/end on `jam.session.events`.
   - Include state transition payload to allow downstream consumers to gate behavior.
   - Why: consistent event stream for read models and gateway fan-out.
   - Alternative considered: no lifecycle events for join/leave. Rejected because downstream state would become implicit.

## Risks / Trade-offs

- **[Risk] Partial lifecycle enforcement (queue guarded, playback not guarded, or vice versa)** -> **Mitigation:** add shared active-session check helper and cross-service integration tests for blocked writes after session end.
- **[Risk] Membership/permission state drift under concurrent joins/leaves** -> **Mitigation:** use atomic repository operations and deterministic idempotent updates where possible.
- **[Risk] Existing clients depend on current `end` behavior without host ownership checks** -> **Mitigation:** document behavior change and validate integration tests for host vs non-host paths.
- **[Trade-off] Additional Redis reads on write path** -> **Mitigation:** keep metadata compact and co-locate keys to minimize lookup overhead.

## Migration Plan

1. Add session metadata/membership repository operations and key schema.
2. Implement lifecycle service methods and wire new HTTP routes (`join`, `leave`).
3. Add ownership checks and state gating in `end`, queue writes, and playback writes.
4. Publish lifecycle events and verify event envelopes/topics.
5. Run service unit/API/integration suites; release behind controlled rollout as needed.
6. Rollback strategy: disable new routes and bypass state gating with feature toggle if critical regression appears.

## Open Questions

- Should session-state write rejections map to `409` (`session_ended`) or `400` (`invalid_state`) across all command paths?
- Is re-join after non-host leave allowed in MVP, or should membership be immutable once removed?
- Should host explicit `end` require premium entitlement in addition to ownership, or ownership alone is sufficient once session exists?
