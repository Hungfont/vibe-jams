## ADDED Requirements

### Requirement: Session lifecycle endpoints
The system SHALL expose Jam session lifecycle endpoints for `create`, `join`, `leave`, and `end`.

#### Scenario: Create session
- **WHEN** an eligible user sends a valid create request
- **THEN** the system creates an active Jam session and returns a created response with session identity and host identity

#### Scenario: Join session
- **WHEN** an authenticated user sends a valid join request for an active Jam session
- **THEN** the system adds the user as a participant and returns a success response

#### Scenario: Leave session
- **WHEN** an authenticated participant sends a leave request for an active Jam session
- **THEN** the system removes that participant and returns a success response

#### Scenario: End session
- **WHEN** an authorized host sends an end request for an active Jam session
- **THEN** the system transitions the session to ended and returns a success response

### Requirement: Premium and ownership authorization policy
The system SHALL require premium entitlement for host session creation and SHALL require host ownership for explicit session end.

#### Scenario: Non-premium create attempt
- **WHEN** an authenticated non-premium user requests session creation
- **THEN** the system returns `403` with `premium_required`

#### Scenario: Non-host end attempt
- **WHEN** an authenticated non-host participant requests explicit session end
- **THEN** the system returns a forbidden ownership error response

### Requirement: Session metadata and permissions persistence
The system SHALL persist session metadata and participant permissions in Redis keyspace for each Jam session.

#### Scenario: Metadata persisted on create
- **WHEN** a Jam session is created
- **THEN** the system stores host identity, session status, and creation metadata in Redis

#### Scenario: Permissions updated on membership change
- **WHEN** participants join or leave
- **THEN** the system updates participant membership and permission records in Redis

### Requirement: Host leave ends session and blocks further writes
The system SHALL end the Jam session when the host leaves and SHALL reject subsequent queue/playback write operations for that session.

#### Scenario: Host leaves active session
- **WHEN** the current host sends a leave request
- **THEN** the system marks the session as ended with host-leave cause

#### Scenario: Write command after session ended
- **WHEN** a queue or playback write command targets a session marked ended
- **THEN** the system rejects the command with a deterministic session-ended error response
