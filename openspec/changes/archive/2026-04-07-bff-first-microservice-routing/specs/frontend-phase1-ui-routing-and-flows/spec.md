## MODIFIED Requirements

### Requirement: Browser calls SHALL use frontend API routes as the only backend access boundary
The frontend SHALL route all browser-initiated requests through App Router API endpoints and MUST NOT call backend service domains directly from browser components. For non-auth microservice HTTP calls, frontend server routes MUST route through api-service BFF first via api-gateway.

#### Scenario: Browser submits protected action
- **WHEN** a user triggers create, join, queue, playback, catalog lookup, or orchestration actions
- **THEN** the browser calls a frontend API route that forwards through `api-gateway -> api-service BFF` before downstream microservice delegation

#### Scenario: Direct downstream call bypassing BFF attempted
- **WHEN** a frontend server route attempts to call jam-service, playback-service, catalog-service, or realtime bootstrap endpoint directly for non-auth flows
- **THEN** the implementation is non-compliant with this requirement

### Requirement: Realtime synchronization SHALL apply monotonic event updates and snapshot fallback recovery
Jam room realtime state SHALL process only monotonic aggregateVersion updates, SHALL ignore stale or duplicate events, and SHALL recover via snapshot when gaps or stale reconnect cursors are detected. Realtime bootstrap configuration MUST be fetched through BFF-first HTTP flow.

#### Scenario: Realtime bootstrap configuration path
- **WHEN** frontend requests realtime ws bootstrap config
- **THEN** frontend route forwards via `api-gateway -> api-service BFF` before returning ws config to client for direct websocket connect

#### Scenario: Gap or stale reconnect
- **WHEN** gap is detected or reconnect cursor is stale
- **THEN** the frontend fetches authoritative session snapshot, replaces local projection, and resumes incremental processing
