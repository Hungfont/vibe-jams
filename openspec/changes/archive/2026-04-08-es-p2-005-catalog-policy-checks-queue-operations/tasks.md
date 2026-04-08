## 1. Catalog Policy Contract and Mapping

- [x] 1.1 Extend catalog lookup contract/adapter to expose deterministic policy status and restriction reason metadata.
- [x] 1.2 Add deterministic jam-service mapping from catalog policy status to queue add outcomes (`track_restricted`, `track_unavailable`, `track_not_found`).
- [x] 1.3 Keep lookup schema compatibility for policy-on and policy-off execution paths.

## 2. Jam Queue Command Integration

- [x] 2.1 Integrate catalog policy checks into jam-service queue add command pre-mutation validation.
- [x] 2.2 Ensure policy rejection happens before queue mutation commit, preserving queueVersion/idempotency semantics.
- [x] 2.3 Add rollout-safe policy toggle handling that preserves baseline behavior when policy checks are disabled.

## 3. Regression and Contract Tests

- [x] 3.1 Add jam-service tests for queue add outcomes: allowed, unavailable, restricted, and not-found.
- [x] 3.2 Add regression tests proving duplicate retry/idempotency behavior is unchanged by policy check rollout.
- [x] 3.3 Add policy-on and policy-off test coverage for deterministic queue consistency behavior.

## 4. Validation and Runbook

- [x] 4.1 Execute targeted backend suites for jam-service/catalog policy integration paths.
- [x] 4.2 Update docs/runbooks/run.md with concise ES-P2-005 test flow and expected deterministic policy outcomes.
- [x] 4.3 Confirm OpenSpec apply readiness and surface approval gate summary for `/opsx:apply`.
