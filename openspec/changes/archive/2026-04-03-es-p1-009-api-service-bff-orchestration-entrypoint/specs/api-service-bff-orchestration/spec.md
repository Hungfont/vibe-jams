## ADDED Requirements

### Requirement: API-service MUST expose MVP BFF orchestration entrypoint
The API-service SHALL expose a BFF orchestration entrypoint for MVP web flows that coordinates auth, jam, playback, and catalog dependencies behind a single client-facing API surface.

#### Scenario: BFF entrypoint handles MVP web request
- **WHEN** a web client sends a valid MVP orchestration request to API-service
- **THEN** API-service routes downstream calls to required auth, jam, playback, and catalog services and returns a unified BFF response

### Requirement: BFF orchestration MUST enforce deterministic timeout and error mapping
The API-service SHALL apply bounded timeout policies for downstream calls and MUST return deterministic normalized error semantics when dependencies fail or exceed timeout budgets.

#### Scenario: Downstream timeout during orchestration
- **WHEN** one required downstream service exceeds configured timeout budget
- **THEN** API-service returns a deterministic dependency-timeout result using the shared BFF error envelope

#### Scenario: Downstream dependency unavailable during orchestration
- **WHEN** one required downstream service is unavailable
- **THEN** API-service returns deterministic dependency-unavailable semantics and does not return ambiguous pass-through errors

### Requirement: BFF response aggregation MUST provide stable MVP contract
The API-service SHALL aggregate downstream auth, jam, playback, and catalog data into a stable response contract required by MVP web client views.

#### Scenario: Successful aggregation for MVP view load
- **WHEN** all required downstream calls succeed
- **THEN** API-service returns aggregated payload fields with stable schema for MVP web client integration

#### Scenario: Partial aggregation under optional dependency failure
- **WHEN** an optional downstream segment fails but required segments succeed
- **THEN** API-service returns deterministic partial-result semantics without violating required response schema fields
