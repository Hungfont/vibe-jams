# api-gateway-authn-middleware Specification

## Purpose
TBD - created by archiving change api-gateway-authn-authz. Update Purpose after archive.
## Requirements
### Requirement: API gateway MUST strip client-supplied X-Auth headers before processing
The api-gateway SHALL remove any `X-Auth-*` headers present in the inbound client request before applying authN middleware, preventing identity spoofing by untrusted clients.

#### Scenario: Client-supplied X-Auth header is stripped
- **WHEN** a client sends a request with an `X-Auth-UserId` or any other `X-Auth-*` header
- **THEN** api-gateway strips all `X-Auth-*` headers before token verification and only injects values derived from its own validation result

### Requirement: API gateway MUST verify Bearer tokens for all non-public routes using the existing auth-service validate endpoint
The api-gateway SHALL extract bearer credentials for protected routes by preferring the `Authorization` header and MAY fallback to trusted auth cookie token fields when header auth is absent. It SHALL call `POST /internal/v1/auth/validate` on auth-service for every request not matching the public route bypass list.

#### Scenario: Valid token on protected route
- **WHEN** a client sends a request to a protected route with a valid bearer token in `Authorization`
- **THEN** api-gateway calls `POST /internal/v1/auth/validate`, receives normalized claims, and proceeds to forward the request

#### Scenario: Cookie token fallback on protected route
- **WHEN** a protected route request has no `Authorization` header but includes a valid `auth_token` or `token` cookie value
- **THEN** api-gateway extracts bearer token from cookie, calls `POST /internal/v1/auth/validate`, and proceeds on success

#### Scenario: Missing credentials on protected route
- **WHEN** a protected route request has neither `Authorization` bearer token nor supported auth cookie token
- **THEN** api-gateway returns `401` with deterministic error code `missing_credentials` and does not forward the request

#### Scenario: Invalid or expired token on protected route
- **WHEN** auth-service `POST /internal/v1/auth/validate` returns `401` for the submitted token
- **THEN** api-gateway returns `401` with deterministic error code `invalid_token` and does not forward the request

### Requirement: API gateway MUST inject normalized claim headers for authenticated requests
After a successful call to `POST /internal/v1/auth/validate`, the api-gateway SHALL inject `X-Auth-UserId`, `X-Auth-Plan`, `X-Auth-Scope`, `X-Auth-SessionState` headers into the upstream request using values from the validate response (`userId`, `plan`, `scope`, `sessionState`).

#### Scenario: Claims injected on successful validation
- **WHEN** `POST /internal/v1/auth/validate` returns a `200` with the normalized claim payload
- **THEN** api-gateway sets `X-Auth-UserId`, `X-Auth-Plan`, `X-Auth-Scope`, `X-Auth-SessionState` on the forwarded request

#### Scenario: Public route receives no gateway-injected X-Auth headers
- **WHEN** a public auth route bypasses authN middleware and is forwarded to auth-service
- **THEN** the forwarded request MUST NOT contain gateway-injected `X-Auth-*` headers

### Requirement: Auth-service validate call MUST use a bounded timeout
The api-gateway SHALL apply a configurable timeout to the `POST /internal/v1/auth/validate` call and treat a timeout as an authentication failure.

#### Scenario: Auth-service validate times out
- **WHEN** auth-service does not respond to the validate call within the configured timeout
- **THEN** api-gateway returns `503` with deterministic error code `auth_service_unavailable` and does not forward the request

#### Scenario: Auth-service validate returns non-200 response
- **WHEN** auth-service returns a non-200 response for the validate call
- **THEN** api-gateway returns `401` with deterministic error code `invalid_token`

