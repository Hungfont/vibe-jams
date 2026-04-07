## MODIFIED Requirements

### Requirement: API-service MUST expose MVP BFF orchestration entrypoint
The API-service SHALL expose a BFF entrypoint surface for MVP web flows where non-auth frontend microservice HTTP calls are mediated by BFF semantics. In addition to orchestration, BFF routes MAY include command/query delegation endpoints that preserve identity and envelope constraints.

#### Scenario: Non-auth frontend call enters BFF surface
- **WHEN** a frontend non-auth HTTP operation is forwarded by api-gateway
- **THEN** api-service BFF processes the call under the same identity and deterministic error mapping constraints used by orchestration

### Requirement: BFF response aggregation MUST provide stable MVP contract
The API-service SHALL aggregate or delegate downstream jam/playback/catalog/realtime-bootstrap data into stable response contracts required by MVP web client views.

#### Scenario: Non-orchestration delegated response shape remains stable
- **WHEN** frontend calls BFF endpoint families beyond orchestration
- **THEN** responses remain deterministic and compatible with frontend envelope mapping expectations
