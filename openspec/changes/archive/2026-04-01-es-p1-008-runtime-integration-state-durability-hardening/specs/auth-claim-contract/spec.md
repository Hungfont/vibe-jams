## ADDED Requirements

### Requirement: Runtime auth claim validation SHALL use real token validation integration
Non-test profiles SHALL validate auth claims through configured runtime token validation integration and MUST NOT rely on static in-memory claim fixtures.

#### Scenario: Valid token through runtime validator
- **WHEN** a request includes valid token in non-test runtime profile
- **THEN** claim extraction uses configured runtime validator and returns normalized claim contract

#### Scenario: Runtime validator unavailable
- **WHEN** non-test runtime cannot reach configured token validation integration
- **THEN** request fails with deterministic unauthorized or dependency-unavailable semantics and protected write is not performed
