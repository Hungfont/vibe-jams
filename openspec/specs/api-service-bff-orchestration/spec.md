# api-service-bff-orchestration Specification

## Purpose
TBD - synchronized from change es-p1-009-api-service-bff-orchestration-entrypoint. Update Purpose after archive.
## Requirements
### Requirement: API-service MUST expose MVP BFF orchestration entrypoint
The API-service SHALL expose a BFF orchestration entrypoint for MVP web flows that coordinates auth identity, jam, and optional catalog dependencies behind a single client-facing API surface. Auth identity MUST be sourced from gateway-forwarded `X-Auth-*` headers (`X-Auth-UserId`, `X-Auth-Plan`, `X-Auth-SessionState`, `X-Auth-Scope`) injected by api-gateway. The API-service MUST NOT call `POST /internal/v1/auth/validate` directly for incoming client requests. The API-service SHALL forward the gateway-injected `X-Auth-*` headers to downstream services (jam-service, playback-service), and the raw `Authorization` header MUST NOT be forwarded.

#### Scenario: BFF entrypoint handles MVP web request with gateway-forwarded identity
- **WHEN** a web client sends a valid orchestration request through api-gateway and api-gateway has injected `X-Auth-*` headers
- **THEN** api-service reads identity from `X-Auth-UserId`, `X-Auth-Plan`, `X-Auth-SessionState`, `X-Auth-Scope` headers, calls jam and optional catalog dependencies, and returns a unified BFF response

#### Scenario: BFF entrypoint rejects request missing gateway identity headers
- **WHEN** an orchestration request arrives at api-service without the `X-Auth-UserId` header
- **THEN** api-service returns `401` with deterministic error code `unauthorized` and MUST NOT process the request

#### Scenario: BFF entrypoint rejects request with non-valid sessionState header
- **WHEN** an orchestration request arrives with `X-Auth-SessionState` value that is not `valid`
- **THEN** api-service returns `401` with deterministic error code `unauthorized`

#### Scenario: X-Auth headers forwarded to downstream dependencies
- **WHEN** api-service makes a downstream request as part of orchestration
- **THEN** downstream requests include gateway-injected `X-Auth-UserId`, `X-Auth-Plan`, `X-Auth-SessionState`, `X-Auth-Scope` headers and MUST NOT include an `Authorization` header

### Requirement: BFF orchestration MUST be side-effect free
The orchestration endpoint MUST execute read aggregation only and MUST NOT trigger playback mutation.

#### Scenario: playbackCommand is provided in orchestration request
- **WHEN** request body contains `playbackCommand`
- **THEN** API-service returns `400` with deterministic error code `invalid_input`
- **AND** no playback mutation is executed

### Requirement: BFF orchestration MUST enforce deterministic timeout and error mapping
The API-service SHALL apply bounded timeout policies for downstream calls and MUST return deterministic normalized error semantics when dependencies fail or exceed timeout budgets.

#### Scenario: Downstream timeout during orchestration
- **WHEN** one required downstream service exceeds configured timeout budget
- **THEN** API-service returns a deterministic dependency-timeout result using the shared BFF error envelope

#### Scenario: Downstream dependency unavailable during orchestration
- **WHEN** one required downstream service is unavailable
- **THEN** API-service returns deterministic dependency-unavailable semantics and does not return ambiguous pass-through errors

### Requirement: BFF response aggregation MUST provide stable MVP contract
The API-service SHALL aggregate downstream jam and optional catalog data into a stable response contract required by MVP web client views.

#### Scenario: Successful aggregation for MVP view load
- **WHEN** all required downstream calls succeed
- **THEN** API-service returns stable orchestration payload without playback execution result

#### Scenario: Partial aggregation under optional dependency failure
- **WHEN** optional catalog dependency fails and required dependencies succeed
- **THEN** API-service returns deterministic partial-result semantics while preserving required fields

