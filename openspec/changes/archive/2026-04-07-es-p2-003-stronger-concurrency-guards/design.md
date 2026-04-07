## Context

Queue reorder/remove operations currently have asymmetric optimistic concurrency behavior: reorder paths require expected queue version while remove paths can still succeed without explicit version intent. Playback command conflict responses also vary by endpoint and do not provide deterministic retry metadata for client reconciliation. Frontend conflict handling currently depends on generic version-conflict error interpretation instead of a shared retry guidance schema.

This change is cross-cutting across jam-service, playback-service, shared response contracts, and frontend reconciliation adapters. Existing OpenSpec capabilities for queue handling, playback pipeline, and frontend jam flows already define version-aware behavior and require additive tightening rather than new capability creation.

## Goals / Non-Goals

**Goals:**
- Enforce mandatory expectedQueueVersion for reorder and remove mutations.
- Standardize 409 conflict responses with a shared retry guidance payload that includes authoritative version metadata.
- Ensure playback state updates include playbackEpoch and queueVersion together for deterministic event/state ordering.
- Preserve backward-compatible response envelopes while strengthening required payload fields for conflict handling.
- Add contract tests across services to verify schema compatibility.

**Non-Goals:**
- Introducing new mutation endpoints or changing route paths.
- Redesigning queue idempotency semantics for add operations.
- Changing realtime transport protocol or websocket bootstrap flow.

## Decisions

1. Shared conflict payload shape for 409 responses
- Decision: Extend conflict error payload to include a retry guidance object with authoritative queue/playback metadata.
- Rationale: Frontend can perform deterministic reconciliation without endpoint-specific parsing.
- Alternatives considered:
  - Per-endpoint custom conflict payloads: rejected due to schema drift risk.
  - Returning only generic message strings: rejected because clients cannot reliably recover.

2. Mandatory expectedQueueVersion for remove and reorder writes
- Decision: Reject reorder/remove requests lacking expectedQueueVersion or carrying stale versions with deterministic 409 payloads.
- Rationale: Ensures optimistic concurrency checks are explicit and symmetric across write paths.
- Alternatives considered:
  - Optional version checks on remove: rejected because stale remove operations can silently override concurrent intent.
  - Server-side implicit snapshot compare only: rejected due to non-deterministic client retries.

3. Playback update contract requires playbackEpoch + queueVersion
- Decision: All accepted playback state update payloads must carry both playbackEpoch and queueVersion.
- Rationale: queueVersion alone cannot disambiguate playback timeline resets or session lifecycle transitions.
- Alternatives considered:
  - Keep playbackEpoch optional: rejected because client ordering remains ambiguous after reconnect/retry.

4. Frontend reconciliation consumes retry guidance schema
- Decision: Frontend queue/playback conflict handlers read standardized retry guidance and trigger snapshot/version reconciliation path.
- Rationale: Reduces duplicated conflict handling logic and aligns with backend contract tests.
- Alternatives considered:
  - Preserve existing generic conflict fallback only: rejected due to inconsistent recovery behavior.

## Risks / Trade-offs

- [Risk] Existing clients that omit expectedQueueVersion on remove may receive new validation/conflict failures. -> Mitigation: update frontend adapters in the same change and document required field.
- [Risk] Conflict payload expansion may diverge between jam-service and playback-service. -> Mitigation: add shared schema tests and service contract tests.
- [Risk] Playback event/state producers could emit incomplete payloads during rollout. -> Mitigation: enforce contract checks in tests and reject incomplete updates at boundary validation.
- [Trade-off] Slightly larger 409 payloads increase response size. -> Mitigation: keep retry guidance minimal and focused on reconciliation fields.
