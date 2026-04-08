## Why

Queue add commands currently validate basic track existence/playability, but they do not model catalog policy restrictions as a first-class gate in the queue command path. ES-P2-005 is needed to ensure restricted or unavailable tracks are rejected deterministically while preserving idempotency and queue-version consistency guarantees.

## What Changes

- Add catalog policy decision checks to jam queue add command handling before mutation commit.
- Extend catalog lookup contract usage so jam queue handlers can read policy status and reason codes required for deterministic command outcomes.
- Keep queue idempotency and optimistic/version consistency semantics unchanged when policy checks are enabled.
- Add rollout-safe behavior for policy-on and policy-off modes to support regression-safe adoption.
- Add regression test coverage for policy-allowed, policy-restricted, policy-unavailable, and policy-disabled paths.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `jams-queue-command-handling`: Queue add command requirements are expanded to include deterministic catalog policy restriction checks without changing idempotency/version behavior.
- `catalog-track-validation-api`: Catalog lookup requirement is expanded to expose policy status and deterministic restriction semantics needed by queue command callers.

## Impact

- Backend services:
  - `backend/jam-service` queue command pipeline, catalog client integration, deterministic error mapping.
  - `backend/catalog-service` lookup response contract/fixtures for policy status fields.
- Shared contracts/tests:
  - Integration and contract tests for queue add outcomes across policy-on and policy-off modes.
- Documentation:
  - `docs/runbooks/run.md` test flow updates for catalog policy queue checks.
