## ADDED Requirements

### Requirement: Runtime token validation SHALL use configured auth validation source
Services consuming auth claims for jam/playback authorization SHALL validate tokens through configured runtime auth validation source and MUST NOT rely on static in-memory claim fixtures in runtime profiles.

#### Scenario: Token validated via runtime auth source
- **WHEN** a request includes an access token in runtime profile
- **THEN** claim normalization uses the configured runtime auth validation source before authorization decisions

### Requirement: Auth validation failures MUST be deterministic
When runtime auth validation source is unavailable or returns invalid payload, services SHALL fail authorization deterministically without mutating protected state.

#### Scenario: Runtime auth validation source unavailable
- **WHEN** token validation is attempted and auth validation source is unavailable
- **THEN** request is rejected with deterministic unauthorized/dependency-unavailable semantics and protected writes are not executed
