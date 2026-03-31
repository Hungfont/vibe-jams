## Why

Phase 1 depends on a reliable event backbone so jam session state, queue changes, playback transitions, and analytics actions can be published consistently across services. Defining Kafka foundations now reduces cross-service drift and unblocks downstream consumers such as realtime fanout.

## What Changes

- Define and provision Phase 1 Kafka topics for session, queue, playback, and analytics event streams with LLD-aligned partitions and retention.
- Introduce producer libraries/adapters in `jam-service`, `playback-service`, and `api-service` that enforce topic selection and keying rules.
- Standardize a shared event envelope contract with required metadata (`eventId`, `eventType`, `sessionId`, `aggregateVersion`, `occurredAt`, `payload`).
- Establish envelope validation and serialization guarantees so consumers can parse events consistently.
- Add verification coverage for keying, envelope schema compliance, and producer publish paths.

## Capabilities

### New Capabilities
- `kafka-event-foundation`: Provides standardized topic configuration, producer behavior, and shared event envelope guarantees for MVP jam and analytics streams.

### Modified Capabilities
- None.

## Impact

- Affected code: `jam-service`, `playback-service`, and `api-service` producer paths plus shared contract package(s).
- APIs: No external REST API changes; internal event contract becomes explicit and validated.
- Dependencies: Kafka cluster provisioning, ACL setup, and shared schema distribution.
- Systems: Realtime and analytics consumers rely on stable topic/key/envelope semantics for ordering and compatibility.
