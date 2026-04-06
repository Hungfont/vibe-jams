## Why

The orchestration endpoint is used in SSR bootstrap and periodic refresh, but current behavior can trigger playback mutation, which violates read-path expectations. A side-effect-free orchestration contract is required to keep retries deterministic and prevent accidental playback actions.

## What Changes

- Restrict orchestration to read aggregation only: auth (required), jam state (required), and catalog lookup (optional/degradable).
- Reject `playbackCommand` in orchestration request with deterministic `400 invalid_input`.
- Ensure playback mutation remains only on the dedicated playback command endpoint.
- Update orchestration contract schema and sequence documentation to match behavior.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `api-service-bff-orchestration`: enforce side-effect-free orchestration and deterministic invalid_input rejection for playbackCommand.

## Impact

- Backend `api-service` orchestration service, DTOs, and OpenAPI schema.
- Frontend orchestration response type.
- Docs: sequence flow and runbook flow.
- Integration tests in `api-service` BFF package.
