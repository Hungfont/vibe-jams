## ADDED Requirements

### Requirement: Auth-service MUST provide stable claim contract for Jam flows
The system SHALL expose a normalized auth claim contract consumed by Jam create and end authorization logic.

#### Scenario: Claim contract contains required identity and plan fields
- **WHEN** auth validation succeeds for a Jam request
- **THEN** the normalized claim payload MUST include `userId`, `plan`, and `sessionState`

#### Scenario: Contract is consumed consistently by Jam entrypoints
- **WHEN** Jam create and Jam end handlers consume auth claims
- **THEN** both entrypoints MUST use the same claim contract type and validation semantics

### Requirement: Claim validation MUST fail fast on missing required fields
The system SHALL treat missing required claim fields as unauthorized context.

#### Scenario: userId missing from claims
- **WHEN** a validated auth payload is missing `userId`
- **THEN** the request MUST be rejected with `401` and `unauthorized`

#### Scenario: plan missing from claims for entitlement check
- **WHEN** a validated auth payload is missing `plan`
- **THEN** the request MUST be rejected with `401` and `unauthorized`

### Requirement: Premium plan mapping MUST be deterministic
The system SHALL map claim plan values to entitlement outcomes in a deterministic manner shared by Jam authorization.

#### Scenario: Plan value maps to premium entitlement
- **WHEN** claim plan value is recognized as premium
- **THEN** entitlement evaluation MUST allow protected Jam operations

#### Scenario: Plan value maps to non-premium entitlement
- **WHEN** claim plan value is recognized as non-premium
- **THEN** entitlement evaluation MUST reject protected Jam operations with `403` and `premium_required`
