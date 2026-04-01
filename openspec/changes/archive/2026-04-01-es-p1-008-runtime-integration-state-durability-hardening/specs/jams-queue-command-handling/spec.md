## ADDED Requirements

### Requirement: Queue runtime state SHALL survive process restart
Queue state, queueVersion, and queue-add idempotency records SHALL be persisted in durable runtime storage in non-test profiles.

#### Scenario: Queue state recovery after restart
- **WHEN** one or more queue mutations are committed and service restarts
- **THEN** snapshot read returns the latest committed queue order and queueVersion from durable storage

#### Scenario: Idempotent retry after restart
- **WHEN** an already-committed queue-add command is retried with the same idempotency key after restart
- **THEN** no duplicate queue item is created and previously committed outcome is returned
