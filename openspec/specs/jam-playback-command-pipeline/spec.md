# jam-playback-command-pipeline Specification

## Purpose
TBD - created by archiving change es-p1-005-playback-command-pipeline. Update Purpose after archive.
## Requirements
### Requirement: Host playback command endpoint
The system SHALL expose a session-scoped playback command endpoint that accepts host-issued playback commands for active jam sessions. Before accepting commands that target or resolve a track, the system SHALL validate track existence and playability through the catalog contract.

#### Scenario: Command accepted for host
- **WHEN** an authenticated host sends a valid playback command to an active session playback command endpoint and catalog validation confirms the referenced track is playable
- **THEN** the system accepts the command and returns an accepted response

#### Scenario: Command rejected for missing track
- **WHEN** a playback command references a `trackId` that catalog resolves as not found
- **THEN** the system rejects the command with deterministic `track_not_found` response and does not execute state transition

#### Scenario: Command rejected for unavailable track
- **WHEN** a playback command references a `trackId` that catalog resolves as unavailable
- **THEN** the system rejects the command with deterministic `track_unavailable` response and does not execute state transition

#### Scenario: Command rejected for ended session
- **WHEN** a playback command is sent for a session marked ended
- **THEN** the system rejects the command with a deterministic session-ended error response

### Requirement: Authorization and host-only enforcement
The system SHALL reject playback commands when authentication is missing/invalid, when session state is not active, or when actor lacks playback control permission for the target session. Host actors are always authorized. Guest actors SHALL require `canControlPlayback=true` in projected permission state.

#### Scenario: Missing or invalid authentication
- **WHEN** a playback command request is sent without valid authentication context
- **THEN** the system returns an unauthorized error response

#### Scenario: Authenticated guest command attempt without permission
- **WHEN** an authenticated guest without playback permission sends a playback command for a session
- **THEN** the system returns a deterministic forbidden permission error response

#### Scenario: Authenticated guest command attempt with permission
- **WHEN** an authenticated guest with playback permission sends a playback command for a session
- **THEN** the system authorizes command execution and applies normal validation/processing semantics

### Requirement: Queue-version stale command protection
The system SHALL validate command staleness against Redis-backed queue version state and reject stale playback commands. The stale-command response SHALL return `409 version_conflict` and include retry guidance payload with authoritative queue and playback metadata for client reconciliation.

#### Scenario: Stale queue version in command
- **WHEN** a playback command includes an expected queue version lower than current queue state
- **THEN** the system rejects the command with `409 version_conflict`, includes retry guidance payload with current queueVersion and playbackEpoch metadata, and does not execute state transition

### Requirement: Playback transition event publication
The system SHALL publish accepted playback transitions to Kafka topic `jam.playback.events` keyed by `sessionId`.

#### Scenario: Successful transition publish
- **WHEN** a playback command is accepted and executed
- **THEN** the system publishes a playback transition event to `jam.playback.events` with key `sessionId`

### Requirement: Command-to-event integration verification
The system SHALL provide integration tests that validate command acceptance/rejection behavior and event publication for the playback command pipeline.

#### Scenario: Integration path validation
- **WHEN** integration tests execute playback command flows for accepted and rejected cases
- **THEN** test assertions confirm accepted commands emit transition events and rejected commands do not emit them

### Requirement: Playback runtime state SHALL be durably persisted
Accepted playback transitions SHALL be committed to durable runtime storage before command completion in non-test profiles.

#### Scenario: Playback transition persists across restart
- **WHEN** a playback transition command is accepted and service restarts
- **THEN** recovered playback state reflects the latest committed transition and sequence metadata

#### Scenario: Durable commit precedes event publication
- **WHEN** a playback transition is accepted in runtime mode
- **THEN** persistent state commit occurs before publishing transition event to Kafka

### Requirement: Playback state update contract includes epoch and queue version
The system SHALL include both `playbackEpoch` and `queueVersion` in playback state update payloads used by command responses and emitted transition state snapshots.

#### Scenario: Accepted command returns epoch and queue version
- **WHEN** a playback command is accepted and returns updated playback state metadata
- **THEN** the payload includes both `playbackEpoch` and `queueVersion`

