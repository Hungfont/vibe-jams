## Why

Concurrent queue and playback writes can still race in edge cases, causing non-deterministic conflict handling between clients and backend services. This change tightens optimistic concurrency contracts so stale writes are rejected consistently and clients receive deterministic recovery guidance.

## What Changes

- Require expectedQueueVersion on all reorder and remove queue mutation writes.
- Standardize 409 conflict responses with retry guidance payload for queue/playback version conflicts.
- Extend playback state update contract to always include playbackEpoch together with queueVersion.
- Align frontend reconciliation handling to consume retry guidance and reconcile local state from authoritative snapshot/version metadata.
- Add contract compatibility tests covering conflict schema and playback update payload requirements.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- jams-queue-command-handling: enforce expectedQueueVersion for reorder/remove writes and deterministic conflict payload.
- jam-playback-command-pipeline: require playback update payloads to include playbackEpoch and queueVersion, and return retry guidance on stale commands.
- frontend-phase1-ui-routing-and-flows: update conflict reconciliation flow to consume retry guidance schema.

## Impact

- Backend jam-service queue mutation handlers, validation logic, error response contracts, and queue contract tests.
- Backend playback-service command validation/response contract and playback state update contract tests.
- Shared API response envelope/types used by frontend and backend conflict handling paths.
- Frontend jam queue/playback API adapters and conflict reconciliation UI behavior.
- Runbook test flows for deterministic stale-write rejection and retry behavior.
