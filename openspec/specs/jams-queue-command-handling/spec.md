# jams-queue-command-handling Specification

## Purpose
TBD - created by archiving change [ES-P1-002].[Phase 1 (MVP)][1]. Update Purpose after archive.
## Requirements
### Requirement: Queue mutations are atomic and versioned
The `jams` service SHALL apply add, remove, and reorder operations as atomic mutations and SHALL increment `queueVersion` by exactly 1 for each successful mutation. The service SHALL reject queue mutation commands when the target Jam session is not active. For add commands, the service SHALL validate track existence and playability using the catalog contract before mutation.

#### Scenario: Successful add increments version once
- **WHEN** a client sends a valid add command for an active jam session queue and catalog validation confirms the track is playable
- **THEN** the service atomically appends the item and returns a snapshot with `queueVersion` incremented by 1

#### Scenario: Add rejected when track does not exist
- **WHEN** a client sends an add command with a `trackId` that catalog resolves as not found
- **THEN** the service rejects the command with deterministic `track_not_found` response and leaves queue state unchanged

#### Scenario: Add rejected when track is unavailable
- **WHEN** a client sends an add command with a `trackId` that catalog resolves as unavailable
- **THEN** the service rejects the command with deterministic `track_unavailable` response and leaves queue state unchanged

#### Scenario: Successful remove increments version once
- **WHEN** a client sends a valid remove command for an existing queue item in an active jam session
- **THEN** the service atomically removes the item and returns a snapshot with `queueVersion` incremented by 1

#### Scenario: Mutation rejected for ended session
- **WHEN** a queue mutation command is sent for a jam session in ended state
- **THEN** the service rejects the command with a deterministic session-ended error response and leaves queue state unchanged

### Requirement: Add command idempotency prevents duplicate retry effects
The `jams` service SHALL support idempotency for queue-add commands using an idempotency key scoped to a jam session so that retried requests do not create duplicate queue entries.

#### Scenario: Duplicate add retry with same idempotency key
- **WHEN** the same add command is retried with the same idempotency key for the same jam session
- **THEN** the service returns the previously committed result and MUST NOT append a second queue item

### Requirement: Reorder requires expected version and detects conflicts
The `jams` service SHALL require clients to provide `expectedQueueVersion` for reorder commands and SHALL reject stale requests with `409 version_conflict`.

#### Scenario: Reorder with stale expected version
- **WHEN** a reorder command includes an `expectedQueueVersion` lower than the current queue version
- **THEN** the service responds with `409 version_conflict` and leaves queue order unchanged

#### Scenario: Reorder with current expected version
- **WHEN** a reorder command includes an `expectedQueueVersion` equal to the current queue version
- **THEN** the service applies the reorder atomically and returns the updated queue with incremented `queueVersion`

### Requirement: Queue snapshot reflects latest committed state
The `jams` service SHALL provide a queue snapshot response containing ordered queue items and current `queueVersion` that represents the latest committed mutation state.

#### Scenario: Read snapshot after concurrent writes
- **WHEN** concurrent mutation commands complete for the same jam session
- **THEN** a subsequent snapshot request returns the deterministic final queue order and the latest `queueVersion`

