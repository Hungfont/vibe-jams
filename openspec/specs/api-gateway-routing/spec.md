# api-gateway-routing Specification

## Purpose
TBD - created by archiving change api-gateway-authn-authz. Update Purpose after archive.
## Requirements
### Requirement: API gateway MUST be the sole external ingress for all client traffic
The api-gateway SHALL receive all inbound client HTTP requests and route them to the appropriate upstream service. No upstream service SHALL be directly reachable from external clients.

#### Scenario: Non-auth request routed to api-service
- **WHEN** a client sends a request to any path that does not match the public auth bypass list
- **THEN** api-gateway applies authN middleware and forwards the request to api-service on success

#### Scenario: Public auth request routed directly to auth-service
- **WHEN** a client sends a request matching a public auth path (`POST /v1/auth/login`, `POST /v1/auth/refresh`, `POST /v1/auth/logout`, `GET /v1/auth/me`)
- **THEN** api-gateway forwards the request directly to auth-service without token verification

### Requirement: API gateway MUST strip the Authorization header before forwarding to api-service
After successful token verification, the api-gateway SHALL remove the `Authorization` header from requests forwarded to api-service and other non-auth upstreams.

#### Scenario: Authorization header stripped on authenticated forward
- **WHEN** a request passes gateway authN and is forwarded to api-service
- **THEN** the upstream request MUST NOT contain the original `Authorization` header

#### Scenario: Public auth routes are forwarded unmodified
- **WHEN** a public auth route request is forwarded to auth-service
- **THEN** the gateway forwards the request as-is, preserving all headers including `Authorization`

### Requirement: API gateway MUST enforce a per-request upstream timeout
The api-gateway SHALL apply a configurable maximum timeout for upstream calls and return a deterministic timeout response if the upstream does not respond within budget.

#### Scenario: Upstream exceeds timeout budget
- **WHEN** an upstream service does not respond within the configured gateway timeout
- **THEN** api-gateway returns `504` with deterministic error code `upstream_timeout`

#### Scenario: Upstream responds within timeout
- **WHEN** an upstream service responds before the timeout budget is exhausted
- **THEN** api-gateway forwards the upstream response status and body to the client unmodified

