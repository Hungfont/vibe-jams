## Why

Queue and playback flows currently rely on service-local checks for track validity, which can drift and produce inconsistent behavior. A shared catalog validation contract is needed now so both `jam-service` and `playback-service` reject missing or unavailable tracks deterministically before state mutation.

## What Changes

- Add a catalog lookup API for `trackId` that returns playable metadata required by queue and playback validations.
- Add deterministic validation outcomes for `track_not_found` and `track_unavailable` cases.
- Integrate catalog pre-checks into `jam-service` queue command path before mutation.
- Integrate catalog pre-checks into `playback-service` command path before transition execution.
- Add contract tests to verify schema compatibility and deterministic error mapping across services.

## Capabilities

### New Capabilities
- `catalog-track-validation-api`: Track metadata lookup and playability contract for command pre-checks.

### Modified Capabilities
- `jams-queue-command-handling`: Queue command requirements updated to require catalog track validation before successful add/mutation paths.
- `jam-playback-command-pipeline`: Playback command requirements updated to require catalog availability validation before accepted transitions.

## Impact

- Affected services: `catalog-service`, `jam-service`, `playback-service`.
- Affected APIs/contracts: catalog track lookup response schema and deterministic error surface for invalid tracks.
- Affected dependencies: shared client/integration contract for catalog lookup across command services.
- Test impact: add/extend contract and integration tests for queue/playback rejection behavior and schema verification.
