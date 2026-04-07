## ADDED Requirements

### Requirement: API-service BFF SHALL perform authoritative host-only checks for delegated jam policy commands
For delegated jam moderation and permission policy routes, api-service BFF SHALL evaluate host-only authorization before forwarding requests to jam-service.

#### Scenario: Non-host moderation command denied at api-service
- **WHEN** non-host actor invokes delegated moderation route
- **THEN** api-service returns deterministic `403 host_only` and does not proxy request downstream

#### Scenario: Host moderation command forwarded by api-service
- **WHEN** host actor invokes delegated moderation route
- **THEN** api-service passes authorization check and proxies request to jam-service

#### Scenario: Non-host permission command denied at api-service
- **WHEN** non-host actor invokes delegated permission route
- **THEN** api-service returns deterministic `403 host_only` and does not proxy request downstream

### Requirement: API-service host-only authorization checks SHALL use normalized gateway identity and jam session state
The api-service authorization guard SHALL evaluate host-only policy decisions using gateway-injected `X-Auth-*` identity headers and jam session state lookup.

#### Scenario: Required identity headers missing for policy route
- **WHEN** delegated policy route is missing required identity headers
- **THEN** api-service returns deterministic unauthorized semantics and does not proxy request

#### Scenario: Jam state lookup unavailable during policy check
- **WHEN** api-service cannot retrieve jam session state for host evaluation
- **THEN** api-service returns deterministic dependency-unavailable semantics and does not proxy request
