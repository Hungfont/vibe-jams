## ADDED Requirements

### Requirement: Queue state SHALL be durable across restart boundaries
The `jams` service SHALL persist queue items, queue version, and idempotency records in durable storage so successful mutations survive restart.

#### Scenario: Queue state recovered after restart
- **WHEN** an add command is committed and the service restarts
- **THEN** queue snapshot returns previously committed item order and `queueVersion` from durable state

### Requirement: Idempotency MUST remain effective after restart
The `jams` service SHALL preserve idempotency key history in durable storage so retried commands after restart do not create duplicate queue mutations.

#### Scenario: Retry add with same idempotency key after restart
- **WHEN** a previously successful add command is retried with the same idempotency key after service restart
- **THEN** the service returns prior result and MUST NOT append a duplicate queue item
