# api-gateway-authn-middleware Specification

## Purpose
TBD - created by archiving change api-gateway-authn-authz. Update Purpose after archive.
## Requirements
### Requirement: API gateway MUST strip client-supplied X-Auth headers before processing
The api-gateway SHALL remove any `X-Auth-*` headers present in the inbound client request before applying authN middleware, preventing identity spoofing by untrusted clients.

#### Scenario: Client-supplied X-Auth header is stripped
- **WHEN** a client sends a request with an `X-Auth-UserId` or any other `X-Auth-*` header
- **THEN** api-gateway strips all `X-Auth-*` headers before token verification and only injects values derived from its own validation result

### Requirement: API gateway MUST verify Bearer tokens for all non-public routes using local JWT verification
The api-gateway SHALL extract bearer credentials for protected routes by preferring the `Authorization` header and MAY fallback to trusted auth cookie token fields when header auth is absent. It SHALL verify tokens locally using shared HS256 `TokenVerifier` for every request not matching the public route bypass list.

#### Scenario: Valid token on protected route
- **WHEN** a client sends a request to a protected route with a valid bearer token in `Authorization`
- **THEN** api-gateway verifies JWT locally, extracts normalized claims, and proceeds to forward the request

#### Scenario: Cookie token fallback on protected route
- **WHEN** a protected route request has no `Authorization` header but includes a valid `auth_token` or `token` cookie value
- **THEN** api-gateway extracts bearer token from cookie, verifies locally, and proceeds on success

#### Scenario: Missing credentials on protected route
- **WHEN** a protected route request has neither `Authorization` bearer token nor supported auth cookie token
- **THEN** api-gateway returns `401` with deterministic error code `missing_credentials` and does not forward the request

#### Scenario: Invalid or expired token on protected route
- **WHEN** local JWT verification fails (invalid signature, expired token, or unknown key id)
- **THEN** api-gateway returns `401` with deterministic error code `invalid_token` and does not forward the request

### Requirement: API gateway MUST inject normalized claim headers for authenticated requests
After successful local JWT verification, the api-gateway SHALL inject `X-Auth-UserId`, `X-Auth-Plan`, `X-Auth-Scope`, `X-Auth-SessionState` headers into the upstream request using normalized claim values (`userId`, `plan`, `scope`, `sessionState`).

#### Scenario: Claims injected on successful validation
- **WHEN** local JWT verification succeeds with a normalized claim payload
- **THEN** api-gateway sets `X-Auth-UserId`, `X-Auth-Plan`, `X-Auth-Scope`, `X-Auth-SessionState` on the forwarded request

#### Scenario: Authorization preserved when token originates from cookie fallback
- **WHEN** api-gateway resolves credentials from `auth_token` or `token` cookie
- **THEN** forwarded request includes `Authorization: Bearer <resolved_token>` and injected `X-Auth-*` headers

#### Scenario: Public route receives no gateway-injected X-Auth headers
- **WHEN** a public auth route bypasses authN middleware and is forwarded to auth-service
- **THEN** the forwarded request MUST NOT contain gateway-injected `X-Auth-*` headers

### Requirement: Gateway JWT verifier configuration MUST be explicitly provided
The api-gateway SHALL require configured JWT verification key material at startup using `AUTH_JWT_ACTIVE_KID` and `AUTH_JWT_ACTIVE_SECRET`, and MAY load rotated keys from `AUTH_JWT_PREVIOUS_KEYS`.

#### Scenario: Missing active JWT key configuration
- **WHEN** `AUTH_JWT_ACTIVE_KID` or `AUTH_JWT_ACTIVE_SECRET` is missing/empty
- **THEN** api-gateway startup fails with deterministic configuration validation error

#### Scenario: Previous key material supports key rotation
- **WHEN** `AUTH_JWT_PREVIOUS_KEYS` contains valid `kid:secret` pairs
- **THEN** tokens signed by those previous keys remain verifiable during rotation window

