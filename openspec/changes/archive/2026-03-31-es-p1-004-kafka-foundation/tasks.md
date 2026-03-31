## 1. Kafka Topic Provisioning Baseline

- [x] 1.1 Define topic configuration for `jam.session.events`, `jam.queue.events`, `jam.playback.events`, and `analytics.user.actions` with LLD-aligned partitions and retention.
- [x] 1.2 Implement topic provisioning flow (or bootstrap script) and ACL bindings for producer/consumer principals.
- [x] 1.3 Add environment validation checks to confirm topic existence and policy correctness.

## 2. Shared Event Envelope Contract

- [x] 2.1 Create shared envelope model with required metadata fields and payload container.
- [x] 2.2 Add envelope validation/serialization helpers used by all producer services.
- [x] 2.3 Document contract rules for session-scoped fields and aggregate version monotonicity expectations.

## 3. Create and initial microservice
- [x] 3.1 Create new folder ./backend/playback-service and setup initial backend project for playback-service.
- [x] 3.2 Create new folder ./backend/api-service and setup initial backend project for api-service.
- [x] 3.3 Create new folder ./backend/auth-service and setup initial backend project for auth-service.
- [x] 3.4 Create new folder ./backend/catalog-service and setup initial backend project for catalog-service.
- [x] 3.5 Create new folder ./backend/playback-service and setup initial backend project for playback-service.
- [x] 3.5 Create new folder ./backend/rt-gateway and setup initial backend project for rt-gateway.

## 4. Producer Integration by Service
- [x] 3.1 Integrate producer adapter in `jam-service` for `jam.session.events` and `jam.queue.events` with `sessionId` keying.
- [x] 3.2 Integrate producer adapter in `playback-service` for `jam.playback.events` with `sessionId` keying.
- [x] 3.3 Integrate producer adapter in `api-service` for `analytics.user.actions` with `userId` keying.
- [x] 3.4 Ensure publish paths fail fast on missing key fields or invalid envelope metadata.

## 5. Verification and Rollout Readiness

- [x] 4.1 Add tests for envelope validation and serialization in shared contract package.
- [x] 4.2 Add service-level tests asserting topic selection and keying rules for each producer.
- [x] 4.3 Add consumer-side contract test proving envelope parse compatibility across all Phase 1 topics.
- [x] 4.4 Add rollout checklist for provisioning order, smoke checks, and rollback switches.
