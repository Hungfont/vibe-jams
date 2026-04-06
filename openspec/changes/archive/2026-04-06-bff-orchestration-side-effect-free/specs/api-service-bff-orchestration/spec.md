## MODIFIED Requirements

### Requirement: BFF orchestration MUST be side-effect free
The orchestration endpoint MUST execute read aggregation only and MUST NOT trigger playback mutation.

#### Scenario: playbackCommand is provided in orchestration request
- **WHEN** request body contains `playbackCommand`
- **THEN** API-service returns `400` with deterministic error code `invalid_input`
- **AND** no playback mutation is executed

### Requirement: API-service MUST expose MVP BFF orchestration entrypoint
The API-service SHALL expose a BFF orchestration entrypoint for MVP web flows that coordinates auth, jam, and optional catalog dependencies behind a single client-facing API surface.

#### Scenario: BFF entrypoint handles MVP web request
- **WHEN** a web client sends a valid orchestration request
- **THEN** API-service calls required auth and jam dependencies, optionally enriches with catalog, and returns a unified BFF response

### Requirement: BFF response aggregation MUST provide stable MVP contract
The API-service SHALL aggregate downstream auth, jam, and optional catalog data into a stable response contract required by MVP web client views.

#### Scenario: Successful aggregation for MVP view load
- **WHEN** all required downstream calls succeed
- **THEN** API-service returns stable orchestration payload without playback execution result

#### Scenario: Partial aggregation under optional dependency failure
- **WHEN** optional catalog dependency fails and required dependencies succeed
- **THEN** API-service returns deterministic partial-result semantics while preserving required fields
