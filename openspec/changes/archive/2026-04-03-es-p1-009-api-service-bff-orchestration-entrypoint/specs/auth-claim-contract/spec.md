## MODIFIED Requirements

### Requirement: Auth-service MUST provide stable claim contract for Jam flows
The system SHALL expose a normalized auth claim contract consumed by Jam create/end authorization logic and API-service BFF orchestration flows.

#### Scenario: Claim contract contains required identity and plan fields
- **WHEN** auth validation succeeds for a Jam request or API-service BFF orchestration request
- **THEN** the normalized claim payload MUST include `userId`, `plan`, and `sessionState`

#### Scenario: Contract is consumed consistently by Jam entrypoints
- **WHEN** Jam create/Jam end handlers and API-service BFF orchestration handlers consume auth claims
- **THEN** all entrypoints MUST use the same claim contract type and validation semantics
