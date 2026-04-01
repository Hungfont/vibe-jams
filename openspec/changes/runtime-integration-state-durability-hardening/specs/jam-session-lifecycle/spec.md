## ADDED Requirements

### Requirement: Session lifecycle state SHALL survive service restarts
The system SHALL persist jam session lifecycle metadata and membership state in durable storage so active sessions can be recovered after process restart.

#### Scenario: Active session recovered after restart
- **WHEN** a session is created and the owning service restarts
- **THEN** subsequent lifecycle reads and write authorization checks use recovered durable session state without resetting session identity

### Requirement: Concurrent lifecycle writes MUST preserve single terminal outcome
The system SHALL enforce consistent terminal session state under concurrent lifecycle commands using durable write semantics.

#### Scenario: Concurrent host leave and explicit end
- **WHEN** host leave and explicit end commands race for the same active session
- **THEN** the system commits exactly one terminal ended state and rejects redundant terminal transition writes
