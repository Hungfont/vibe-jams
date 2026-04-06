# jam-moderation-controls Specification

## Purpose
TBD - created by archiving change es-p2-002-moderation-controls. Update Purpose after archive.
## Requirements
### Requirement: Jam-service SHALL expose host moderation commands
The system SHALL provide host-only moderation command endpoints for `mute` and `kick` actions scoped by session.

#### Scenario: Host mutes participant
- **WHEN** host submits valid mute command for an active session participant
- **THEN** system marks target participant as muted and returns updated session participant projection

#### Scenario: Host kicks participant
- **WHEN** host submits valid kick command for an active session participant
- **THEN** system removes target participant from active participant projection and returns updated session participant projection

#### Scenario: Non-host moderation attempt
- **WHEN** non-host actor submits moderation command
- **THEN** system returns deterministic `403 host_only`

### Requirement: Moderation actions SHALL publish auditable moderation events
The system SHALL publish moderation events to Kafka topic `jam.moderation.events` using the shared event envelope and moderation audit payload.

#### Scenario: Mute audit payload published
- **WHEN** mute command succeeds
- **THEN** event payload includes `action=mute`, `actorUserId`, `targetUserId`, `reason`, and `occurredAt`

#### Scenario: Kick audit payload published
- **WHEN** kick command succeeds
- **THEN** event payload includes `action=kick`, `actorUserId`, `targetUserId`, `reason`, and `occurredAt`

### Requirement: Muted or kicked users SHALL be blocked from moderated actions
The system SHALL reject blocked jam command actions for users in muted or kicked moderation states.

#### Scenario: Muted participant tries blocked queue mutation
- **WHEN** muted participant issues blocked queue command
- **THEN** command is rejected with deterministic moderation-blocked error and state remains unchanged

#### Scenario: Kicked participant tries blocked queue mutation
- **WHEN** kicked participant issues blocked queue command
- **THEN** command is rejected with deterministic moderation-blocked error and state remains unchanged

### Requirement: Moderation updates SHALL be visible in room realtime flow
Realtime fanout SHALL broadcast moderation events and updated session state transitions to room subscribers.

#### Scenario: Moderation event consumed by gateway
- **WHEN** rt-gateway consumes moderation event for a session
- **THEN** connected room subscribers receive moderation update through websocket outbound event stream

#### Scenario: Reconnect snapshot includes moderation state
- **WHEN** client reconnects and requires snapshot recovery
- **THEN** session snapshot payload includes participant moderation state required for room rendering

