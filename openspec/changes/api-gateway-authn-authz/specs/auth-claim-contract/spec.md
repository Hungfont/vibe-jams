## ADDED Requirements

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
