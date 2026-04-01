## MODIFIED Requirements

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
