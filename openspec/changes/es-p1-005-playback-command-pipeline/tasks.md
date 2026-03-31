## 1. API Contract and Routing

- [x] 1.1 Define playback command request/response models and validation rules (`command`, `clientEventId`, `expectedQueueVersion`).
- [x] 1.2 Add session-scoped playback command endpoint routing and OpenAPI contract in `playback-service`.
- [x] 1.3 Standardize error mapping for invalid input, unauthorized, host-only forbidden, and version conflict responses.

## 2. Authorization and Queue Version Validation

- [x] 2.1 Integrate auth claim validation for playback command requests.
- [x] 2.2 Enforce host-only command policy for target jam session.
- [x] 2.3 Add Redis queue snapshot/version read path used by playback command validation.
- [x] 2.4 Reject stale commands when `expectedQueueVersion` does not match current queue version.

## 3. Command Execution and Event Emission

- [x] 3.1 Implement internal command executor for accepted playback commands (`play`, `pause`, `next`, `prev`, `seek`).
- [x] 3.2 Produce playback transition events to `jam.playback.events` keyed by `sessionId` using shared envelope conventions.
- [x] 3.3 Ensure aggregate version progression and event payload fields are consistent for accepted transitions.

## 4. Integration and Verification

- [x] 4.1 Add integration tests for successful host command acceptance and Kafka publish path.
- [x] 4.2 Add integration tests for unauthorized and non-host command rejection paths.
- [x] 4.3 Add integration tests for stale command rejection (`409 version_conflict`) and non-emission of playback events.
- [x] 4.4 Run service tests and regression checks for related jam/playback contracts.
