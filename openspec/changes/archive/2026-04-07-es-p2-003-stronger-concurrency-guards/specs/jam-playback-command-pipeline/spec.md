## ADDED Requirements

### Requirement: Playback state update contract includes epoch and queue version
The system SHALL include both `playbackEpoch` and `queueVersion` in playback state update payloads used by command responses and emitted transition state snapshots.

#### Scenario: Accepted command returns epoch and queue version
- **WHEN** a playback command is accepted and returns updated playback state metadata
- **THEN** the payload includes both `playbackEpoch` and `queueVersion`

## MODIFIED Requirements

### Requirement: Queue-version stale command protection
The system SHALL validate command staleness against Redis-backed queue version state and reject stale playback commands. The stale-command response SHALL return `409 version_conflict` and include retry guidance payload with authoritative queue and playback metadata for client reconciliation.

#### Scenario: Stale queue version in command
- **WHEN** a playback command includes an expected queue version lower than current queue state
- **THEN** the system rejects the command with `409 version_conflict`, includes retry guidance payload with current queueVersion and playbackEpoch metadata, and does not execute state transition
