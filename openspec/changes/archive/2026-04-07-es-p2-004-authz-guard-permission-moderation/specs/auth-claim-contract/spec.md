## ADDED Requirements

### Requirement: Auth claim fields for policy authorization SHALL be consumed consistently by centralized guard
The centralized api-service policy authorization guard SHALL consume normalized auth claim fields (`userId`, `plan`, `sessionState`, and optional `scope`) consistently for permission and moderation command authorization decisions.

#### Scenario: Valid claims are mapped into authorization decision context
- **WHEN** permission or moderation command entrypoint receives validated auth claims
- **THEN** api-service centralized authorization guard receives consistent claim-derived actor context for policy decision evaluation

#### Scenario: Missing required claim identity fields fails policy authorization deterministically
- **WHEN** required claim identity fields are missing for policy authorization
- **THEN** api-service policy command entrypoint fails deterministically with unauthorized or host-only semantics before downstream forwarding
