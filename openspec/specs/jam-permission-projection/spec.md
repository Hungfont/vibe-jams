# jam-permission-projection Specification

## Purpose
TBD - created by archiving change es-p2-001-permission-projection. Update Purpose after archive.
## Requirements
### Requirement: Jam session permission projection SHALL store granular guest control capabilities
The system SHALL maintain session-scoped guest permission projection fields `canControlPlayback`, `canReorderQueue`, and `canChangeVolume` with deterministic defaults for active sessions.

#### Scenario: Session initializes permission defaults
- **WHEN** a new jam session is created
- **THEN** permission projection is initialized with deterministic default values for guest capabilities

#### Scenario: Session snapshot includes permission projection
- **WHEN** a client or dependent service reads session state snapshot
- **THEN** response includes current permission projection values required for command gating and UI rendering

### Requirement: Host SHALL update permission projection through deterministic command endpoints
The system SHALL expose host-only permission update commands that mutate one or more guest capability flags for a target active session.

#### Scenario: Host updates guest playback control capability
- **WHEN** host submits valid permission update setting `canControlPlayback`
- **THEN** system persists updated projection and returns updated permission state

#### Scenario: Non-host attempts permission update
- **WHEN** non-host actor submits permission update command
- **THEN** system rejects request with deterministic `403 host_only` and leaves projection unchanged

### Requirement: Permission update events SHALL be published and auditable
Accepted permission updates SHALL publish session-scoped permission events to Kafka topic `jam.permission.events` using the shared event envelope.

#### Scenario: Permission update event publish on success
- **WHEN** host permission update command is accepted
- **THEN** system publishes `jam.permission.updated` event with actor identity, changed flags, and session aggregate metadata

#### Scenario: Failed permission update does not publish event
- **WHEN** permission update command is rejected
- **THEN** no `jam.permission.events` record is produced for the rejected mutation

