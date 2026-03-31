## Why

Phase 1 needs an authoritative playback command path so host actions become consistent, ordered playback updates for all participants. This is needed now because Kafka topic foundations are already in place, but command acceptance, version checks, and end-to-end command-to-event behavior are not yet defined and implemented as a single contract.

## What Changes

- Add a playback command API contract for host-issued commands (`play`, `pause`, `next`, `prev`, `seek`) scoped to a jam session.
- Add playback command execution flow in `playback-service` that validates authorization and stale command conditions before transition.
- Add queue-version synchronization checks using Redis-backed queue state to reject stale commands with conflict errors.
- Publish accepted playback transitions to `jam.playback.events` using session-keyed event envelopes.
- Add integration tests covering HTTP command acceptance/rejection and command-to-Kafka publish behavior.

## Capabilities

### New Capabilities
- `jam-playback-command-pipeline`: Host playback command ingestion, authorization/staleness validation, and Kafka transition publishing for jam sessions.

### Modified Capabilities
- None.

## Impact

- Affected services: `backend/playback-service` (primary), `backend/rt-gateway` (downstream playback fanout consumer behavior validation).
- Affected dependencies: Redis queue snapshot/version reads and Kafka topic `jam.playback.events`.
- Affected API surface: playback command endpoint for jam sessions and its error contract for unauthorized/stale command cases.
- Affected testing: integration tests for command acceptance, error mapping, and event publish path.
