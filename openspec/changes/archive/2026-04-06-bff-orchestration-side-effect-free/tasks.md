## 1. Backend Orchestration Behavior

- [x] 1.1 Reject orchestration payloads containing `playbackCommand` with deterministic `400 invalid_input`
- [x] 1.2 Remove playback mutation execution from orchestration flow and keep read aggregation only
- [x] 1.3 Update backend orchestration DTO contract and OpenAPI schema to remove playback output

## 2. Contract and Documentation Alignment

- [x] 2.1 Update frontend orchestration response type to remove playback field
- [x] 2.2 Update sequence documentation to reflect read-only orchestration and dedicated playback mutation endpoint
- [x] 2.3 Update runbook and canonical orchestration spec to match side-effect-free behavior

## 3. Validation

- [x] 3.1 Update integration tests for success and optional-degradation paths without playback mutation
- [x] 3.2 Add integration test for playbackCommand rejection with `400 invalid_input`
- [x] 3.3 Run api-service BFF tests and module tests to confirm behavior
