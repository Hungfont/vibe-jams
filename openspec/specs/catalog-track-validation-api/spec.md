# catalog-track-validation-api Specification

## Purpose
TBD - created by archiving change es-p1-007-catalog-track-validation-api. Update Purpose after archive.
## Requirements
### Requirement: Catalog MUST provide track lookup with playable metadata
The `catalog-service` SHALL expose a `trackId` lookup contract that returns deterministic track existence, playable availability metadata, and policy status metadata for command pre-check use cases.

#### Scenario: Lookup returns playable and policy-allowed track metadata
- **WHEN** a client requests lookup for an existing, playable, and policy-allowed `trackId`
- **THEN** the response includes `trackId`, `isPlayable=true`, and required metadata fields used by queue/playback validation

#### Scenario: Lookup returns unavailable track metadata
- **WHEN** a client requests lookup for an existing but unavailable `trackId`
- **THEN** the response includes `trackId`, `isPlayable=false`, and an unavailability reason code

#### Scenario: Lookup returns policy-restricted track metadata
- **WHEN** a client requests lookup for an existing track that is restricted by catalog policy
- **THEN** the response includes deterministic policy status and restriction reason metadata required for caller-side deterministic `track_restricted` mapping

#### Scenario: Lookup preserves deterministic shape when policy checks are disabled by caller
- **WHEN** a caller uses lookup contract while policy checks are disabled in command service
- **THEN** lookup payload still exposes stable policy metadata fields so caller behavior can be toggled without schema branching

### Requirement: Catalog MUST expose deterministic missing-track outcome
The `catalog-service` SHALL return a deterministic not-found result for unknown `trackId` values.

#### Scenario: Unknown track id lookup
- **WHEN** a client requests lookup for a `trackId` that does not exist
- **THEN** the response maps to deterministic `track_not_found` semantics for callers

### Requirement: Contract schema MUST remain stable across integrations
The catalog track validation response SHALL be covered by contract tests consumed by `jam-service`, `playback-service`, and `api-service` BFF orchestration flows.

#### Scenario: Contract test validates response schema
- **WHEN** contract tests run for catalog lookup integration
- **THEN** required fields and deterministic error semantics match the shared contract expected by command services and API-service BFF orchestration

### Requirement: Runtime catalog validation SHALL use real dependency integration
Catalog validation in non-test profiles SHALL resolve track lookup through configured runtime catalog integration and MUST NOT use in-memory-only source.

#### Scenario: Non-test lookup with real catalog integration
- **WHEN** runtime catalog lookup is requested in non-test profile
- **THEN** lookup is executed through configured catalog dependency integration

#### Scenario: Catalog dependency unavailable in runtime mode
- **WHEN** non-test runtime cannot reach configured catalog dependency
- **THEN** caller receives deterministic dependency-unavailable semantics

