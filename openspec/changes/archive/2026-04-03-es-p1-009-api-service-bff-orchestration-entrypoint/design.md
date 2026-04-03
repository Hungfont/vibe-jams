## Context

Phase 1 services exist as separate APIs for auth, jam lifecycle/queue, playback, and catalog validation. MVP web clients currently carry orchestration logic across these dependencies, which creates repeated client-side composition and inconsistent timeout/error handling.

The change introduces an API-service BFF orchestration entrypoint that centralizes downstream routing and aggregation behavior for MVP web flows while preserving existing service ownership boundaries.

## Goals / Non-Goals

**Goals:**
- Provide a single BFF-facing entrypoint for MVP web integration across auth, jam, playback, and catalog concerns.
- Define deterministic orchestration behavior for downstream call ordering, timeout budgets, and error mapping.
- Ensure claim context from auth validation is propagated consistently through orchestrated downstream requests.
- Define response aggregation structure for MVP web views that require combined jam, playback, and catalog data.

**Non-Goals:**
- Replacing existing domain service endpoints.
- Introducing new domain authorization policies beyond claim propagation and existing auth contracts.
- Implementing recommendation, moderation, or non-MVP aggregation experiences.

## Decisions

1. BFF entrypoint owns orchestration, not domain logic
- Decision: Implement orchestration handlers in API-service that route to existing downstream service contracts and aggregate responses.
- Rationale: Keeps domain services as source of truth while reducing client complexity.
- Alternative considered: direct web-client fan-out to all services. Rejected due to duplicated client logic and inconsistent reliability behavior.

2. Deterministic timeout and error normalization at BFF boundary
- Decision: Use bounded per-dependency timeout policies and normalize dependency failures into a stable response envelope.
- Rationale: Prevents inconsistent client behavior and improves debuggability.
- Alternative considered: pass-through upstream statuses without normalization. Rejected because behavior diverges by dependency and complicates client handling.

3. Claim propagation contract from auth validation to downstream protected routes
- Decision: Preserve normalized claim context fields and propagate actor identity through BFF orchestration calls.
- Rationale: Aligns auth behavior between direct service entrypoints and BFF flows.
- Alternative considered: revalidate independently in each downstream call without shared context propagation. Rejected because it introduces duplication and drift risk.

4. Minimal MVP aggregation contract
- Decision: Aggregate only data required for MVP web screens and keep payload shape stable with explicit partial-failure semantics.
- Rationale: Limits coupling and keeps rollout tractable.
- Alternative considered: broad generic aggregation framework. Rejected as over-scope for Phase 1.

## Risks / Trade-offs

- [Risk] API-service becomes a latency bottleneck for combined flows. -> Mitigation: bounded fan-out, timeout budgets, and endpoint-level latency instrumentation.
- [Risk] Failure mapping may hide useful upstream diagnostics. -> Mitigation: structured internal logs and trace correlation while returning deterministic client-safe errors.
- [Risk] Tight coupling to downstream contracts can create churn during service evolution. -> Mitigation: keep BFF adapters thin and contract-test orchestration boundaries.
- [Risk] Aggregated responses may return partial data under dependency degradation. -> Mitigation: define explicit partial-failure fields and fallback behavior in spec scenarios.

## Migration Plan

1. Add BFF orchestration endpoint and adapter clients in API-service behind MVP route flag.
2. Introduce contract tests for auth claim propagation and downstream error normalization.
3. Integrate MVP web client against BFF endpoint while retaining legacy direct calls as fallback during rollout window.
4. Enable BFF route by environment, monitor latency/error metrics, then make BFF path default.
5. Rollback by disabling BFF route flag and reverting client traffic to direct service paths.

## Open Questions

- What is the exact MVP endpoint surface (single composite endpoint vs small set of task-oriented BFF endpoints)?
- Which response fields are mandatory for first web-screen load vs optional deferred hydration?
- Should partial-failure responses include per-dependency status map in MVP or only normalized summary?
