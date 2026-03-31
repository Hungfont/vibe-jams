## MODIFIED Requirements

### Requirement: Host playback command endpoint
The system SHALL expose a session-scoped playback command endpoint that accepts host-issued playback commands for active jam sessions.

#### Scenario: Command accepted for host
- **WHEN** an authenticated host sends a valid playback command to an active session playback command endpoint
- **THEN** the system accepts the command and returns an accepted response

#### Scenario: Command rejected for ended session
- **WHEN** a playback command is sent for a session marked ended
- **THEN** the system rejects the command with a deterministic session-ended error response

### Requirement: Authorization and host-only enforcement
The system SHALL reject playback commands when authentication is missing/invalid, when the actor is not authorized as host for the target session, or when session state is not active.

#### Scenario: Missing or invalid authentication
- **WHEN** a playback command request is sent without valid authentication context
- **THEN** the system returns an unauthorized error response

#### Scenario: Authenticated non-host command attempt
- **WHEN** an authenticated user without host permission sends a playback command for a session
- **THEN** the system returns a forbidden host-only error response
