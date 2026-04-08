## MODIFIED Requirements

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
