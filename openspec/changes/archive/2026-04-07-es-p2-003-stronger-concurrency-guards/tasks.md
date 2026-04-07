## 1. Shared Contract and Error Schema

- [x] 1.1 Define shared conflict-error payload shape with retry guidance fields for queue/playback version conflicts.
- [x] 1.2 Update API/DTO response contracts so 409 responses include deterministic retry guidance metadata.

## 2. Jam Queue Concurrency Guard Updates

- [x] 2.1 Require expectedQueueVersion for reorder and remove command writes in jam-service request validation.
- [x] 2.2 Reject stale or missing expectedQueueVersion deterministically with 409 conflict payload and preserve queue state.
- [x] 2.3 Add or update queue contract tests for reorder/remove stale-write rejection and conflict schema compatibility.

## 3. Playback Contract Tightening

- [x] 3.1 Update playback command/state update contract to require playbackEpoch together with queueVersion.
- [x] 3.2 Return conflict retry guidance on playback stale command rejection.
- [x] 3.3 Add playback contract tests validating epoch+queueVersion payload compatibility.

## 4. Frontend Version Reconciliation Support

- [x] 4.1 Update frontend jam queue/playback adapters to parse shared retry guidance payload in 409 responses.
- [x] 4.2 Implement deterministic reconciliation path using authoritative queueVersion and playbackEpoch before retry.
- [x] 4.3 Add frontend contract/adapter tests for retry guidance mapping and stale-write recovery behavior.

## 5. Validation and Runbook Alignment

- [x] 5.1 Execute targeted backend and frontend tests covering concurrency guard behavior and contract compatibility.
- [x] 5.2 Update docs/runbooks/run.md with concise test flows for stale-write rejection and playback epoch/version verification.
