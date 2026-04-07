## MODIFIED Requirements

### Requirement: Jam-service SHALL expose host moderation commands
The system SHALL provide host-only moderation command endpoints for `mute` and `kick` actions scoped by session. For delegated frontend route flows, api-service SHALL enforce host-only moderation authorization before forwarding to jam-service, and jam-service SHALL retain host-only guard checks as defense-in-depth.

#### Scenario: Host mutes participant
- **WHEN** host submits valid mute command for an active session participant
- **THEN** system marks target participant as muted and returns updated session participant projection

#### Scenario: Host kicks participant
- **WHEN** host submits valid kick command for an active session participant
- **THEN** system removes target participant from active participant projection and returns updated session participant projection

#### Scenario: Non-host moderation attempt
- **WHEN** non-host actor submits moderation command through delegated route flow
- **THEN** api-service returns deterministic `403 host_only` and does not forward moderation command to jam-service

## ADDED Requirements

### Requirement: Moderation policy audit trail SHALL include actor metadata for accepted and denied decisions
Moderation policy audit trail SHALL remain compatible with actor identity metadata fields while upstream denial behavior is moved to api-service.

#### Scenario: Accepted moderation action records actor metadata
- **WHEN** moderation command authorization succeeds and action executes
- **THEN** audit payload records actor identity metadata with accepted outcome

#### Scenario: Denied moderation action records actor metadata
- **WHEN** moderation command is denied by api-service host-only guard
- **THEN** response is auditable with actor metadata context and no moderation state mutation occurs
