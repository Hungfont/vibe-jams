## MODIFIED Requirements

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
