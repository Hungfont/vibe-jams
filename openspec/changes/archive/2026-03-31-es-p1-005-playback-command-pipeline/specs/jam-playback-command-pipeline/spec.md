## ADDED Requirements

### Requirement: Host playback command endpoint
The system SHALL expose a session-scoped playback command endpoint that accepts host-issued playback commands for jam sessions.

#### Scenario: Command accepted for host
- **WHEN** an authenticated host sends a valid playback command to the session playback command endpoint
- **THEN** the system accepts the command and returns an accepted response

### Requirement: Authorization and host-only enforcement
The system SHALL reject playback commands when authentication is missing/invalid or when the actor is not authorized as host for the target session.

#### Scenario: Missing or invalid authentication
- **WHEN** a playback command request is sent without valid authentication context
- **THEN** the system returns an unauthorized error response

#### Scenario: Authenticated non-host command attempt
- **WHEN** an authenticated user without host permission sends a playback command for a session
- **THEN** the system returns a forbidden host-only error response

### Requirement: Queue-version stale command protection
The system SHALL validate command staleness against Redis-backed queue version state and reject stale playback commands.

#### Scenario: Stale queue version in command
- **WHEN** a playback command includes an expected queue version lower than current queue state
- **THEN** the system rejects the command with a version-conflict error response

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
