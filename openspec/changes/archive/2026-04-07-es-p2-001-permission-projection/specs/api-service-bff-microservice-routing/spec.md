## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: Delegated protected command routes SHALL enforce permission-aware centralized authorization before forwarding
For delegated protected command routes (`playback`, `queue reorder`, `volume`) api-service SHALL consume centralized authZ decision context and permission projection state before forwarding command execution downstream.

#### Scenario: Guest command denied when capability is disabled
- **WHEN** delegated protected command is submitted by guest actor with corresponding capability disabled
- **THEN** api-service rejects request with deterministic forbidden permission error and does not proxy command downstream

#### Scenario: Guest command allowed when capability is enabled
- **WHEN** delegated protected command is submitted by guest actor with corresponding capability enabled
- **THEN** api-service forwards command to downstream service for normal processing