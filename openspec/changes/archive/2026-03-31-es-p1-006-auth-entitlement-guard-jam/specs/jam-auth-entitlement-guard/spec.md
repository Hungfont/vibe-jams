## ADDED Requirements

### Requirement: Jam endpoints MUST enforce auth entitlement guard
The system SHALL apply a shared auth entitlement guard to Jam session create and end flows before business logic execution.

#### Scenario: Guard executes before Jam create
- **WHEN** a client calls the Jam create endpoint
- **THEN** the auth entitlement guard MUST validate auth context before creating a Jam session

#### Scenario: Guard executes before Jam end
- **WHEN** a client calls the Jam end endpoint
- **THEN** the auth entitlement guard MUST validate auth context before ending a Jam session

### Requirement: Unauthorized auth states MUST map to 401 unauthorized
The system SHALL return HTTP `401` with error code `unauthorized` when token or session validation fails.

#### Scenario: Missing bearer token
- **WHEN** a Jam create or end request has no bearer token
- **THEN** the API MUST return `401` and `unauthorized`

#### Scenario: Invalid or expired token
- **WHEN** auth token validation fails due to invalid or expired token
- **THEN** the API MUST return `401` and `unauthorized`

#### Scenario: Invalid session state
- **WHEN** token is present but session state is invalid or revoked
- **THEN** the API MUST return `401` and `unauthorized`

### Requirement: Non-premium users MUST map to 403 premium_required
The system SHALL enforce premium entitlement policy for protected Jam operations and return HTTP `403` with error code `premium_required` when entitlement is insufficient.

#### Scenario: Authenticated non-premium user requests protected Jam action
- **WHEN** a valid authenticated user without premium entitlement calls a protected Jam create or end operation
- **THEN** the API MUST return `403` and `premium_required`

#### Scenario: Premium user requests protected Jam action
- **WHEN** a valid authenticated user with premium entitlement calls a protected Jam create or end operation
- **THEN** the request MUST proceed to the Jam business handler
