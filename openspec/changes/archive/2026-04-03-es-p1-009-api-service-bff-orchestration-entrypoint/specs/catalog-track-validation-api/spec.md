## MODIFIED Requirements

### Requirement: Contract schema MUST remain stable across integrations
The catalog track validation response SHALL be covered by contract tests consumed by `jam-service`, `playback-service`, and `api-service` BFF orchestration flows.

#### Scenario: Contract test validates response schema
- **WHEN** contract tests run for catalog lookup integration
- **THEN** required fields and deterministic error semantics match the shared contract expected by command services and API-service BFF orchestration
