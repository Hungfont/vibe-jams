## MODIFIED Requirements

### Requirement: Realtime synchronization SHALL apply monotonic event updates and snapshot fallback recovery
Jam room realtime state SHALL process only monotonic aggregateVersion updates, SHALL ignore stale or duplicate events, and SHALL recover via snapshot when gaps or stale reconnect cursors are detected. Frontend realtime bootstrap and websocket connect path MUST route through gateway/BFF and MUST NOT connect directly to rt-gateway public URL.

#### Scenario: Monotonic event stream
- **WHEN** incoming event version equals local version plus one
- **THEN** the event is applied to room projection and local version advances

#### Scenario: Gap or stale reconnect
- **WHEN** gap is detected or reconnect cursor is stale
- **THEN** the frontend fetches authoritative session snapshot, replaces local projection, and resumes incremental processing

#### Scenario: Frontend realtime websocket connect path uses gateway/BFF proxy
- **WHEN** frontend receives websocket bootstrap config
- **THEN** frontend connects using gateway/BFF websocket route and does not call rt-gateway `/ws` directly
