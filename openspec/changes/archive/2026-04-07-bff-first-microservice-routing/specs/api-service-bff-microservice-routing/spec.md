## ADDED Requirements

### Requirement: API-service BFF MUST be the mandatory non-auth microservice HTTP hop for frontend flows
For frontend-originated non-auth microservice HTTP operations, api-service BFF SHALL be the first internal service hop after api-gateway and SHALL delegate to jam-service, playback-service, catalog-service, or rt-gateway control endpoints as needed.

#### Scenario: Jam command/query routed via BFF
- **WHEN** frontend sends non-auth jam actions (create, join, leave, end, queue, moderation) through frontend API routes
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
