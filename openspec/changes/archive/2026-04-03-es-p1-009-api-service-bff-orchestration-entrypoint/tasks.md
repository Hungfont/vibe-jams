## 1. API-service BFF Entrypoint Surface

- [x] 1.1 Add MVP BFF orchestration route(s) and request/response envelope contract in api-service
- [x] 1.2 Implement downstream client wiring for auth-service, jam-service, playback-service, and catalog-service
- [x] 1.3 Add endpoint-level config for timeout budgets and dependency routing defaults

## 2. Orchestration and Aggregation Behavior

- [x] 2.1 Implement orchestration handler logic for auth, jam, playback, and catalog call coordination
- [x] 2.2 Implement deterministic aggregation payload mapping for MVP web client view requirements
- [x] 2.3 Implement deterministic partial-result behavior for optional dependency degradation paths

## 3. Error and Timeout Normalization

- [x] 3.1 Add standardized dependency-timeout and dependency-unavailable error mapping in BFF handlers
- [x] 3.2 Add stable upstream-to-BFF error translation for auth, jam, playback, and catalog failures
- [x] 3.3 Add structured telemetry fields for dependency failures and timeout outcomes

## 4. Auth Claim Propagation and Catalog Contract Alignment

- [x] 4.1 Propagate normalized auth claim context through BFF orchestration to protected downstream calls
- [x] 4.2 Add contract tests validating claim contract usage consistency between jam entrypoints and BFF flows
- [x] 4.3 Add contract tests validating catalog lookup response schema compatibility for BFF orchestration consumers

## 5. Integration Validation and Readiness

- [x] 5.1 Add integration tests for successful MVP orchestration request across auth, jam, playback, and catalog
- [x] 5.2 Add integration tests for timeout/failure normalization and partial-result behavior
- [x] 5.3 Update runbook with concise BFF orchestration test flow and expected outcomes
