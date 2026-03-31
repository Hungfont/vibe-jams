## Context

`playback-service` currently contains only scaffold runtime wiring and a Kafka producer adapter for `jam.playback.events`. The MVP story requires a complete host-command path from HTTP command ingestion to validated transition publishing. Existing queue consistency patterns already exist in `jam-service` (version-conflict semantics and Redis-backed queue snapshots), and Kafka foundation work is complete.

Constraints:
- Preserve session ordering via Kafka keying by `sessionId`.
- Reuse stable error semantics already used by jam APIs (`unauthorized`, host-only forbidden, `version_conflict`).
- Keep Phase 1 scope limited to command acceptance and transition publication without Phase 2 features such as advanced epoch policy.

Stakeholders:
- Playback and jam backend engineers implementing command handling.
- Realtime gateway consumers relying on ordered playback updates.
- Client applications expecting deterministic command acceptance/rejection.

## Goals / Non-Goals

**Goals:**
- Define a single authoritative playback command endpoint and execution pipeline in `playback-service`.
- Enforce authorization and stale-command rejection before publishing playback transitions.
- Align transition events with queue-version synchronization and existing event envelope conventions.
- Define integration-level verification for command-to-event behavior.

**Non-Goals:**
- Re-architect queue command ownership in `jam-service`.
- Introduce new Kafka topics or partition/retention changes.
- Deliver Phase 2+ capabilities (`playbackEpoch` conflict rules, advanced reconciliation policies).
- Define client UI reconciliation behavior beyond established event version semantics.

## Decisions

### 1) Endpoint and ownership
Playback commands are handled by `playback-service` at a session-scoped endpoint, consistent with LLD contract direction.  
Rationale: command execution and playback transition publication are domain-owned by playback.  
Alternative considered: route through `api-service` as orchestrator. Rejected for MVP because it adds an extra hop and duplicates command validation concerns.

### 2) Command payload shape for MVP
The command request includes command type and optional staleness guard (`expectedQueueVersion`) plus a client correlation identifier (`clientEventId`) for observability/idempotency extension.  
Rationale: queue-version checks are required by story DoD and LLD sequence, while correlation allows traceability in tests and logs.  
Alternative considered: no explicit expected version in request. Rejected because stale detection becomes implicit and less testable.

### 3) Authorization and entitlement checks
Command execution validates bearer claims and host permission before transition acceptance. Unauthorized requests return `401`; authenticated non-host requests return `403` host-only error.  
Rationale: story requires unauthorized command handling and Jam policies are host-led in MVP.  
Alternative considered: allow guests with toggles. Rejected for Phase 1 scope.

### 4) Queue-version synchronization strategy
Before applying a command transition, the service reads queue snapshot metadata from Redis and compares current version with request guard. On mismatch, command is rejected as `409 version_conflict`.  
Rationale: prevents stale command application when queue changed concurrently.  
Alternative considered: eventual conflict detection in consumers. Rejected because command acceptance must be deterministic at write time.

### 5) Event emission contract
Accepted transitions are emitted to `jam.playback.events` with `sessionId` key and standard envelope fields (`eventId`, `eventType`, `sessionId`, `aggregateVersion`, `occurredAt`, payload).  
Rationale: preserves ordered per-session playback stream and compatibility with downstream consumers.  
Alternative considered: embed full queue snapshot in playback payload. Rejected to keep event size and coupling minimal in MVP.

## Risks / Trade-offs

- **[Cross-service state drift]** Queue version state and playback transition state can diverge if command flow is not atomic enough.  
  → **Mitigation:** fail fast on stale guard mismatch and publish only after successful validation and transition commit.

- **[Contract ambiguity between services]** Endpoint and error semantics may drift from jam-service conventions.  
  → **Mitigation:** reuse existing error code vocabulary and align OpenAPI documentation for command responses.

- **[Idempotency gap]** Duplicate host command submissions may produce duplicate events in MVP.  
  → **Mitigation:** include `clientEventId` in request/event metadata and add follow-up hardening task if strict dedupe is required.

- **[Operational blind spots]** Missing structured traces make stale rejection debugging hard.  
  → **Mitigation:** include request correlation and rejection reason in logs and integration assertions.

## Migration Plan

1. Introduce playback command HTTP contract and handler in `playback-service` behind service-local rollout controls if needed.
2. Wire authorization validator, queue-state reader, command executor, and Kafka producer into one command pipeline.
3. Add integration tests for accepted host commands and rejected unauthorized/stale commands.
4. Validate emitted events on `jam.playback.events` and confirm downstream consumption compatibility.
5. Roll back by disabling command endpoint route or reverting to prior no-op behavior while leaving Kafka foundation unchanged.

## Open Questions

- Should `expectedQueueVersion` be required for all command types in Phase 1, or only for commands that advance track position (`next`, `prev`, `seek`)?
- Is `clientEventId` treated as observability-only in MVP, or should partial idempotency behavior be committed now?
- Does playback aggregate version increment strictly per accepted command, including no-op commands (for example `play` when already playing)?
