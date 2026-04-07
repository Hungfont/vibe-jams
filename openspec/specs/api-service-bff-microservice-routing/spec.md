# api-service-bff-microservice-routing Specification

## Purpose
TBD - created by archiving change bff-first-microservice-routing. Update Purpose after archive.
## Requirements
### Requirement: API-service BFF MUST be the mandatory non-auth microservice HTTP hop for frontend flows
For frontend-originated non-auth microservice HTTP operations, api-service BFF SHALL be the first internal service hop after api-gateway and SHALL delegate to jam-service, playback-service, catalog-service, or rt-gateway control endpoints as needed.

#### Scenario: Jam command/query routed via BFF
- **WHEN** frontend sends non-auth jam actions (create, join, leave, end, queue, moderation, permissions) through frontend API routes
- **THEN** request flow is `frontend API route -> api-gateway -> api-service BFF -> jam-service`

#### Scenario: Playback command routed via BFF
- **WHEN** frontend sends playback command through frontend API routes
- **THEN** request flow is `frontend API route -> api-gateway -> api-service BFF -> playback-service`

#### Scenario: Catalog lookup routed via BFF
- **WHEN** frontend requests track lookup through frontend API routes
- **THEN** request flow is `frontend API route -> api-gateway -> api-service BFF -> catalog-service`

### Requirement: BFF MUST preserve gateway-validated identity context for downstream calls
For delegated non-auth flows, BFF SHALL consume gateway-injected `X-Auth-*` identity headers and forward normalized identity context to downstream dependencies, without forwarding raw `Authorization` headers.

#### Scenario: Downstream identity propagation
- **WHEN** BFF delegates to jam-service or playback-service
- **THEN** BFF forwards `X-Auth-UserId`, `X-Auth-Plan`, `X-Auth-SessionState`, `X-Auth-Scope` and MUST NOT forward `Authorization`

### Requirement: Realtime bootstrap HTTP configuration MUST be routed through BFF
Frontend realtime bootstrap configuration retrieval SHALL be handled through BFF before websocket connect, while websocket transport remains direct to rt-gateway.

#### Scenario: Realtime bootstrap config via BFF
- **WHEN** frontend requests realtime ws config
- **THEN** request flow is `frontend API route -> api-gateway -> api-service BFF -> rt-gateway config endpoint` and frontend receives normalized ws config payload

### Requirement: Delegated protected command routes SHALL enforce permission-aware centralized authorization before forwarding
For delegated protected command routes (`playback`, `queue reorder`, `volume`) api-service SHALL consume centralized authZ decision context and permission projection state before forwarding command execution downstream.

#### Scenario: Guest command denied when capability is disabled
- **WHEN** delegated protected command is submitted by guest actor with corresponding capability disabled
- **THEN** api-service rejects request with deterministic forbidden permission error and does not proxy command downstream

#### Scenario: Guest command allowed when capability is enabled
- **WHEN** delegated protected command is submitted by guest actor with corresponding capability enabled
- **THEN** api-service forwards command to downstream service for normal processing

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

