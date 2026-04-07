## MODIFIED Requirements

### Requirement: BFF response aggregation MUST provide stable MVP contract
The API-service SHALL aggregate or delegate downstream jam/playback/catalog/realtime-bootstrap data into stable response contracts required by MVP web client views, and its published OpenAPI document MUST reflect these supported route families.

#### Scenario: OpenAPI includes delegated BFF route families
- **WHEN** api-service publishes `/swagger/openapi.json`
- **THEN** the specification enumerates orchestration and delegated BFF paths used by frontend BFF-first routing
