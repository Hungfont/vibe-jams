# api-service-bff-orchestration Specification

## Purpose
TBD - synchronized from change es-p1-009-api-service-bff-orchestration-entrypoint. Update Purpose after archive.
## Requirements
### Requirement: API-service MUST expose MVP BFF orchestration entrypoint
The API-service SHALL expose a BFF entrypoint surface for MVP web flows where non-auth frontend microservice HTTP calls are mediated by BFF semantics. In addition to orchestration, BFF routes MAY include command/query delegation endpoints that preserve identity and envelope constraints.

#### Scenario: Non-auth frontend call enters BFF surface
- **WHEN** a frontend non-auth HTTP operation is forwarded by api-gateway
- **THEN** api-service BFF processes the call under the same identity and deterministic error mapping constraints used by orchestration

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
The API-service SHALL aggregate or delegate downstream jam/playback/catalog/realtime-bootstrap data into stable response contracts required by MVP web client views.

#### Scenario: Non-orchestration delegated response shape remains stable
- **WHEN** frontend calls BFF endpoint families beyond orchestration
- **THEN** responses remain deterministic and compatible with frontend envelope mapping expectations

