## ADDED Requirements

### Requirement: Runtime catalog lookup SHALL use durable catalog source
In runtime profiles, catalog track validation SHALL resolve lookups from configured durable catalog storage and MUST NOT rely on in-memory-only seed adapters.

#### Scenario: Runtime lookup source uses configured persistent backend
- **WHEN** `catalog-service` starts in runtime profile with valid persistent backend configuration
- **THEN** track lookup requests are served from the configured durable source

### Requirement: Lookup behavior MUST remain deterministic on dependency failures
Catalog lookup integration SHALL return deterministic service-dependency failure semantics when durable backend connectivity is unavailable.

#### Scenario: Persistent backend unavailable
- **WHEN** catalog lookup is requested while durable backend connectivity is unavailable
- **THEN** the service responds with deterministic dependency-unavailable error semantics for callers
