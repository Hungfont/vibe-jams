## Context

`jam-service` currently validates queue add requests against catalog track existence and playable availability before committing queue mutations. ES-P2-005 introduces catalog policy decisions (for example, region/license restrictions) as an additional gate in queue add handling. The rollout must preserve deterministic command outcomes, existing idempotency semantics, and queue version consistency guarantees.

This change spans multiple backend modules:
- `jam-service` queue add command path and error mapping
- `catalog-service` lookup policy metadata contract
- integration/contract tests and runbook flows

## Goals / Non-Goals

**Goals:**
- Add deterministic catalog policy checks to queue add operations in `jam-service`.
- Expose policy status from catalog lookup contract so queue handlers can make deterministic accept/reject decisions.
- Preserve existing queue idempotency and optimistic consistency behavior.
- Support policy-on and policy-off execution modes for safe rollout and regression testing.

**Non-Goals:**
- Changing queue reorder/remove semantics.
- Changing playback command policy behavior in this change.
- Introducing new user-facing queue APIs beyond existing command surfaces.

## Decisions

1. Queue add command remains the single enforcement point for catalog policy checks.
- Decision: enforce policy in `jam-service` add-command path before mutation commit.
- Rationale: preserves a single command gate and avoids duplicating policy checks in multiple entrypoints.
- Alternative considered: enforce at upstream API-service only. Rejected because direct/internal command paths would diverge and consistency guarantees become harder to enforce.

2. Catalog lookup response will include explicit policy status fields.
- Decision: extend lookup contract consumed by command services to include deterministic policy outcome and reason metadata.
- Rationale: queue handlers require structured policy decision data, not inference from generic availability flags.
- Alternative considered: infer restriction from `isPlayable=false` only. Rejected because unavailable vs restricted semantics must remain distinguishable.

3. Rollout will be controlled by deterministic policy-check toggle.
- Decision: preserve existing behavior when policy checks are disabled; enforce restricted-track rejection only when enabled.
- Rationale: enables safe staged rollout and regression comparison against baseline queue behavior.
- Alternative considered: always-on enforcement immediately. Rejected due to migration risk and reduced rollback flexibility.

4. Queue consistency behavior remains unchanged by policy checks.
- Decision: policy rejection happens before queue mutation and version increment; accepted adds keep current idempotency/version flow.
- Rationale: DoD requires no regressions in queue consistency logic.
- Alternative considered: post-mutation compensation on policy failure. Rejected because it introduces unnecessary complexity and consistency risk.

## Risks / Trade-offs

- [Risk] Catalog policy metadata drift between `catalog-service` and `jam-service` contract usage. -> Mitigation: add contract tests that assert required policy fields and deterministic mappings.
- [Risk] Toggle misconfiguration between environments causes inconsistent behavior. -> Mitigation: document rollout expectations and add explicit policy-on/policy-off test flows in runbook.
- [Risk] New rejection path may accidentally affect idempotent retry semantics. -> Mitigation: add regression tests for duplicate retry behavior under policy-on and policy-off modes.
- [Risk] Ambiguous policy reason codes can lead to inconsistent error mapping. -> Mitigation: define deterministic mapping table in service-level tests and enforce stable error codes.
