## ADDED Requirements

### Requirement: Issued access tokens SHALL map to stable scope-aware claim contract
The system SHALL derive normalized claims from issued access tokens and SHALL include scope semantics used by downstream authorization guards.

#### Scenario: Scope included in normalized claims
- **WHEN** access token is issued through login or refresh flow and later validated
- **THEN** normalized claim contract includes `userId`, `plan`, `sessionState`, and `scope`

#### Scenario: Missing required scope semantics
- **WHEN** token payload is missing required scope metadata for normalized claim construction
- **THEN** validation fails with deterministic unauthorized semantics
