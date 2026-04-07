# auth-claim-contract Specification

## Purpose
TBD - created by archiving change es-p1-006-auth-entitlement-guard-jam. Update Purpose after archive.
## Requirements
### Requirement: Auth-service MUST provide stable claim contract for Jam flows
The system SHALL expose a normalized auth claim contract consumed by Jam create/end authorization logic and API-service BFF orchestration flows.

#### Scenario: Claim contract contains required identity and plan fields
- **WHEN** auth validation succeeds for a Jam request or API-service BFF orchestration request
- **THEN** the normalized claim payload MUST include `userId`, `plan`, and `sessionState`

#### Scenario: Contract is consumed consistently by Jam entrypoints
- **WHEN** Jam create/Jam end handlers and API-service BFF orchestration handlers consume auth claims
- **THEN** all entrypoints MUST use the same claim contract type and validation semantics

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

### Requirement: Runtime auth claim validation SHALL use real token validation integration
Non-test profiles SHALL validate auth claims through configured runtime token validation integration and MUST NOT rely on static in-memory claim fixtures.

#### Scenario: Valid token through runtime validator
- **WHEN** a request includes valid token in non-test runtime profile
- **THEN** claim extraction uses configured runtime validator and returns normalized claim contract

#### Scenario: Runtime validator unavailable
- **WHEN** non-test runtime cannot reach configured token validation integration
- **THEN** request fails with deterministic unauthorized or dependency-unavailable semantics and protected write is not performed

### Requirement: Issued access tokens SHALL map to stable scope-aware claim contract
The system SHALL derive normalized claims from issued access tokens and SHALL include scope semantics used by downstream authorization guards.

#### Scenario: Scope included in normalized claims
- **WHEN** access token is issued through login or refresh flow and later validated
- **THEN** normalized claim contract includes `userId`, `plan`, `sessionState`, and `scope`

#### Scenario: Missing required scope semantics
- **WHEN** token payload is missing required scope metadata for normalized claim construction
- **THEN** validation fails with deterministic unauthorized semantics

### Requirement: Claim field names MUST be consistent across all internal and gateway surfaces
The normalized claim contract fields (`userId`, `plan`, `scope`, `sessionState`) returned by `POST /internal/v1/auth/validate` SHALL map directly to gateway-injected header names (`X-Auth-UserId`, `X-Auth-Plan`, `X-Auth-Scope`, `X-Auth-SessionState`). All downstream services consuming these headers MUST treat these values as equivalent to the validated claim contract.

#### Scenario: Gateway header values match validate response fields
- **WHEN** api-gateway successfully calls `POST /internal/v1/auth/validate` and injects headers
- **THEN** `X-Auth-UserId` value equals `userId`, `X-Auth-Plan` equals `plan`, `X-Auth-SessionState` equals `sessionState`, and `X-Auth-Scope` equals `scope` from the validate response

#### Scenario: Downstream service receives consistent identity context
- **WHEN** api-service or another downstream service reads `X-Auth-UserId` from a gateway-forwarded request
- **THEN** the value MUST be the same `userId` that was validated by auth-service for that request

### Requirement: sessionState MUST be validated as active before forwarding claims
The api-gateway SHALL reject requests where the validate response contains `sessionState` that is not `valid`. A non-`valid` sessionState indicates a revoked or invalid session and MUST NOT result in claim injection.

#### Scenario: Non-valid sessionState is treated as unauthorized
- **WHEN** `POST /internal/v1/auth/validate` returns `200` but `sessionState` is not `valid`
- **THEN** api-gateway returns `401` with deterministic error code `invalid_token` and does not forward the request

### Requirement: Auth claim fields for policy authorization SHALL be consumed consistently by centralized guard
The centralized api-service policy authorization guard SHALL consume normalized auth claim fields (`userId`, `plan`, `sessionState`, and optional `scope`) consistently for permission and moderation command authorization decisions.

#### Scenario: Valid claims are mapped into authorization decision context
- **WHEN** permission or moderation command entrypoint receives validated auth claims
- **THEN** api-service centralized authorization guard receives consistent claim-derived actor context for policy decision evaluation

#### Scenario: Missing required claim identity fields fails policy authorization deterministically
- **WHEN** required claim identity fields are missing for policy authorization
- **THEN** api-service policy command entrypoint fails deterministically with unauthorized or host-only semantics before downstream forwarding

