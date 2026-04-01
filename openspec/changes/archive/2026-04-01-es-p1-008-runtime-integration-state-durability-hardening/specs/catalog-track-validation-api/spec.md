## ADDED Requirements

### Requirement: Runtime catalog validation SHALL use real dependency integration
Catalog validation in non-test profiles SHALL resolve track lookup through configured runtime catalog integration and MUST NOT use in-memory-only source.

#### Scenario: Non-test lookup with real catalog integration
- **WHEN** runtime catalog lookup is requested in non-test profile
- **THEN** lookup is executed through configured catalog dependency integration

#### Scenario: Catalog dependency unavailable in runtime mode
- **WHEN** non-test runtime cannot reach configured catalog dependency
- **THEN** caller receives deterministic dependency-unavailable semantics
