## ADDED Requirements

### Requirement: Session lifecycle state SHALL be restart-recoverable
Session metadata and participant permission state SHALL be loaded from durable runtime storage after process restart in non-test profiles.

#### Scenario: Active session metadata recovered after restart
- **WHEN** an active session exists and service restarts
- **THEN** subsequent lifecycle read and write decisions use recovered session metadata from durable storage
