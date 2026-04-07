## MODIFIED Requirements

### Requirement: Queue UI SHALL enforce version-aware mutation and deterministic conflict recovery
Queue add, remove, and reorder interactions SHALL execute through frontend routes, SHALL honor queue version behavior, and SHALL provide conflict recovery with snapshot refresh path. Conflict handling SHALL consume shared retry guidance payload from `409 version_conflict` responses.

#### Scenario: Queue reorder conflict
- **WHEN** reorder response indicates `version_conflict`
- **THEN** the UI reads retry guidance metadata, refetches authoritative snapshot, and offers retry without corrupting local projection

#### Scenario: Track validation error
- **WHEN** add flow encounters `track_not_found` or `track_unavailable`
- **THEN** the UI preserves existing queue state and surfaces explicit, actionable error feedback

### Requirement: Playback UI SHALL enforce host-only policy and deterministic command outcome handling
Playback controls SHALL only permit host command execution and SHALL map unauthorized, host-only, version conflict, and session-ended outcomes to deterministic UI states.

#### Scenario: Host command accepted
- **WHEN** host submits a valid playback command with required fields
- **THEN** UI shows accepted pending state and waits for realtime authoritative transition

#### Scenario: Non-host command attempt
- **WHEN** non-host user attempts playback command
- **THEN** command is blocked with host-only reason and no state transition is applied

#### Scenario: Playback version conflict with retry guidance
- **WHEN** playback command response indicates `version_conflict`
- **THEN** the UI uses retry guidance fields including queueVersion and playbackEpoch to reconcile local state before enabling retry
