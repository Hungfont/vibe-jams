## MODIFIED Requirements

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
