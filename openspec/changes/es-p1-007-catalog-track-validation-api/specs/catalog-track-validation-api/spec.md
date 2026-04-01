## ADDED Requirements

### Requirement: Catalog MUST provide track lookup with playable metadata
The `catalog-service` SHALL expose a `trackId` lookup contract that returns deterministic track existence and playable availability metadata for command pre-check use cases.

#### Scenario: Lookup returns playable track metadata
- **WHEN** a client requests lookup for an existing and playable `trackId`
- **THEN** the response includes `trackId`, `isPlayable=true`, and required metadata fields used by queue/playback validation

#### Scenario: Lookup returns unavailable track metadata
- **WHEN** a client requests lookup for an existing but unavailable `trackId`
- **THEN** the response includes `trackId`, `isPlayable=false`, and an unavailability reason code

### Requirement: Catalog MUST expose deterministic missing-track outcome
The `catalog-service` SHALL return a deterministic not-found result for unknown `trackId` values.

#### Scenario: Unknown track id lookup
- **WHEN** a client requests lookup for a `trackId` that does not exist
- **THEN** the response maps to deterministic `track_not_found` semantics for callers

### Requirement: Contract schema MUST remain stable across integrations
The catalog track validation response SHALL be covered by contract tests consumed by `jam-service` and `playback-service`.

#### Scenario: Contract test validates response schema
- **WHEN** contract tests run for catalog lookup integration
- **THEN** required fields and deterministic error semantics match the shared contract expected by command services
