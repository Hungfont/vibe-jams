## 1. Catalog Validation Contract

- [ ] 1.1 Define track lookup response schema for `trackId`, `isPlayable`, and deterministic not-found/unavailable outcomes
- [ ] 1.2 Implement `catalog-service` lookup endpoint for `trackId` with deterministic `track_not_found` and `track_unavailable` semantics
- [ ] 1.3 Add unit tests for catalog lookup behavior across playable, unavailable, and missing tracks

## 2. Shared Integration Adapter

- [ ] 2.1 Implement shared catalog client/adapter contract for command services with timeout and fail-fast behavior
- [ ] 2.2 Map catalog outcomes to deterministic command-layer errors (`track_not_found`, `track_unavailable`) in reusable helper logic
- [ ] 2.3 Add contract tests to validate request/response schema compatibility between catalog and clients

## 3. Jam Service Queue Pre-check Integration

- [ ] 3.1 Integrate catalog pre-check into `jam-service` queue add command path before state mutation
- [ ] 3.2 Ensure rejected tracks do not mutate queue state and return deterministic errors
- [ ] 3.3 Add integration tests verifying add accept/reject behavior and no-mutation guarantees for invalid/unavailable tracks

## 4. Playback Service Pre-check Integration

- [ ] 4.1 Integrate catalog pre-check into `playback-service` command acceptance flow before transition execution
- [ ] 4.2 Ensure invalid/unavailable tracks are rejected deterministically and no playback transition event is published
- [ ] 4.3 Add integration tests for host command accept/reject with missing/unavailable track conditions

## 5. Rollout and Verification

- [ ] 5.1 Add feature toggle/config path for enabling catalog validation checks in queue and playback command flows
- [ ] 5.2 Define rollout and rollback checklist with monitoring for reject rates and command latency impact
- [ ] 5.3 Execute end-to-end contract/integration test suite and record validation evidence for definition of done
