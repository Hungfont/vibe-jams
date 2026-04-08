# jams-queue-command-handling Specification

## Purpose
TBD - created by archiving change [ES-P1-002].[Phase 1 (MVP)][1]. Update Purpose after archive.
## Requirements
### Requirement: Queue mutations are atomic and versioned
The `jams` service SHALL apply add, remove, and reorder operations as atomic mutations and SHALL increment `queueVersion` by exactly 1 for each successful mutation. The service SHALL reject queue mutation commands when the target Jam session is not active. For add commands, the service SHALL validate track existence, playable availability, and catalog policy restriction status using the catalog contract before mutation. Reorder and remove commands SHALL require `expectedQueueVersion` and SHALL reject requests that omit it.

#### Scenario: Successful add increments version once
- **WHEN** a client sends a valid add command for an active jam session queue and catalog validation confirms the track is playable and policy-allowed
- **THEN** the service atomically appends the item and returns a snapshot with `queueVersion` incremented by 1

#### Scenario: Add rejected when track does not exist
- **WHEN** a client sends an add command with a `trackId` that catalog resolves as not found
- **THEN** the service rejects the command with deterministic `track_not_found` response and leaves queue state unchanged

#### Scenario: Add rejected when track is unavailable
- **WHEN** a client sends an add command with a `trackId` that catalog resolves as unavailable
- **THEN** the service rejects the command with deterministic `track_unavailable` response and leaves queue state unchanged

#### Scenario: Add rejected when track is policy-restricted
- **WHEN** a client sends an add command with a `trackId` that catalog resolves as policy-restricted while policy checks are enabled
- **THEN** the service rejects the command with deterministic `track_restricted` response and leaves queue state unchanged

#### Scenario: Add fallback preserves baseline behavior when policy checks are disabled
- **WHEN** a client sends an add command for a track that is policy-restricted and policy checks are disabled
- **THEN** queue add handling follows baseline existence/playability behavior without introducing new version or idempotency side effects

#### Scenario: Successful remove increments version once
- **WHEN** a client sends a valid remove command for an existing queue item in an active jam session with `expectedQueueVersion` equal to the current queue version
- **THEN** the service atomically removes the item and returns a snapshot with `queueVersion` incremented by 1

#### Scenario: Remove rejected when expected version missing
- **WHEN** a remove command omits `expectedQueueVersion`
- **THEN** the service rejects the command deterministically and leaves queue state unchanged

#### Scenario: Mutation rejected for ended session
- **WHEN** a queue mutation command is sent for a jam session in ended state
- **THEN** the service rejects the command with a deterministic session-ended error response and leaves queue state unchanged

### Requirement: Add command idempotency prevents duplicate retry effects
The `jams` service SHALL support idempotency for queue-add commands using an idempotency key scoped to a jam session so that retried requests do not create duplicate queue entries.

#### Scenario: Duplicate add retry with same idempotency key
- **WHEN** the same add command is retried with the same idempotency key for the same jam session
- **THEN** the service returns the previously committed result and MUST NOT append a second queue item

### Requirement: Reorder requires expected version and detects conflicts
The `jams` service SHALL require clients to provide `expectedQueueVersion` for reorder commands and SHALL reject stale requests with `409 version_conflict`. The conflict response SHALL include retry guidance payload containing authoritative queue metadata needed to reconcile and retry.

#### Scenario: Reorder with stale expected version
- **WHEN** a reorder command includes an `expectedQueueVersion` lower than the current queue version
- **THEN** the service responds with `409 version_conflict`, includes retry guidance payload with current queue version metadata, and leaves queue order unchanged

#### Scenario: Reorder with current expected version
- **WHEN** a reorder command includes an `expectedQueueVersion` equal to the current queue version
- **THEN** the service applies the reorder atomically and returns the updated queue with incremented `queueVersion`

#### Scenario: Remove with stale expected version
- **WHEN** a remove command includes an `expectedQueueVersion` lower than the current queue version
- **THEN** the service responds with `409 version_conflict`, includes retry guidance payload with current queue version metadata, and leaves queue unchanged

### Requirement: Queue snapshot reflects latest committed state
The `jams` service SHALL provide a queue snapshot response containing ordered queue items and current `queueVersion` that represents the latest committed mutation state.

#### Scenario: Read snapshot after concurrent writes
- **WHEN** concurrent mutation commands complete for the same jam session
- **THEN** a subsequent snapshot request returns the deterministic final queue order and the latest `queueVersion`

### Requirement: Queue runtime state SHALL survive process restart
Queue state, queueVersion, and queue-add idempotency records SHALL be persisted in durable runtime storage in non-test profiles.

#### Scenario: Queue state recovery after restart
- **WHEN** one or more queue mutations are committed and service restarts
- **THEN** snapshot read returns the latest committed queue order and queueVersion from durable storage

#### Scenario: Idempotent retry after restart
- **WHEN** an already-committed queue-add command is retried with the same idempotency key after restart
- **THEN** no duplicate queue item is created and previously committed outcome is returned

### Requirement: Queue reorder authorization SHALL enforce projected guest permissions
The `jams` service command handling path SHALL allow reorder commands from host actors and from guest actors only when `canReorderQueue=true` in session permission projection state.

#### Scenario: Guest reorder denied when capability disabled
- **WHEN** guest actor submits reorder command while `canReorderQueue` is disabled
- **THEN** command is rejected with deterministic forbidden permission error and queue state remains unchanged

#### Scenario: Guest reorder allowed when capability enabled
- **WHEN** guest actor submits reorder command while `canReorderQueue` is enabled and optimistic version checks pass
- **THEN** reorder mutation is applied with normal version increment semantics

