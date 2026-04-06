# kafka-event-foundation Specification

## Purpose
TBD - created by archiving change es-p1-004-kafka-foundation. Update Purpose after archive.
## Requirements
### Requirement: Phase 1 Kafka topics are provisioned with defined policies
The system SHALL provision `jam.session.events`, `jam.queue.events`, `jam.playback.events`, `jam.moderation.events`, and `analytics.user.actions` with partition and retention settings aligned to defined runtime baseline.

#### Scenario: Topic provisioning matches expected settings
- **WHEN** infrastructure provisioning completes for an environment
- **THEN** each required topic exists and its partition and retention configuration matches the defined values

### Requirement: Producers publish with deterministic topic keying
The system SHALL publish jam session, queue, playback, and moderation events using `sessionId` as the Kafka key and SHALL publish analytics events using `userId` as the Kafka key.

#### Scenario: Jam event uses session key
- **WHEN** `jam-service` or `playback-service` publishes a jam-related event
- **THEN** the produced record key equals the event `sessionId`

#### Scenario: Analytics event uses user key
- **WHEN** `api-service` publishes an analytics action event
- **THEN** the produced record key equals the acting `userId`

### Requirement: Event envelope contract is standardized across producers
All producer services SHALL emit events using a shared envelope containing `eventId`, `eventType`, `sessionId` (if session-scoped), `aggregateVersion` (if session-scoped), `occurredAt`, and `payload`.

#### Scenario: Producer emits valid envelope fields
- **WHEN** a producer serializes an outbound event
- **THEN** required envelope fields are present and serialized in the shared contract format

### Requirement: Invalid envelope data is rejected before publish
Producer libraries SHALL validate envelope and key inputs and MUST reject publish attempts that do not satisfy required contract fields.

#### Scenario: Missing required metadata blocks publish
- **WHEN** a producer receives an event missing required envelope metadata or required keying value
- **THEN** the publish call fails with a validation error and no Kafka record is produced

### Requirement: Consumers can parse envelope consistently
The system SHALL provide parse/validation compatibility so consumers can decode any event from the defined Phase 1 topics using the shared envelope contract.

#### Scenario: Consumer validation passes for producer events
- **WHEN** a consumer receives events from any Phase 1 topic produced by supported services
- **THEN** the consumer validation path successfully parses envelope metadata and payload structure without service-specific branching

### Requirement: Runtime transport MUST use real Kafka in non-test profiles
Services participating in jam queue, playback, and realtime fanout SHALL initialize real Kafka clients in non-test profiles and MUST reject noop transport configuration.

#### Scenario: Non-test runtime with real broker configuration
- **WHEN** jam-service, playback-service, or rt-gateway starts in a non-test profile with valid broker settings
- **THEN** required Kafka producer and consumer clients are initialized and runtime is marked ready

#### Scenario: Non-test runtime with noop transport configured
- **WHEN** a non-test profile attempts to start with noop or mock Kafka transport
- **THEN** startup fails with deterministic invalid-runtime-adapter error and service readiness stays false

### Requirement: Local profile SHALL provide deterministic broker fallback
The system SHALL apply deterministic localhost broker defaults in local profile when explicit broker endpoints are not provided.

#### Scenario: Local profile without explicit broker endpoint
- **WHEN** a local runtime starts without service-specific broker endpoint configuration
- **THEN** startup uses documented localhost broker fallback and logs effective configuration source

### Requirement: Shared noop transport SHALL be reused for non-default Kafka services
Services that do not use Kafka as default transport SHALL reuse shared `NoOpsProducer` and `NoOpsConsumer` from the shared module instead of service-local noop implementations.

#### Scenario: Service boot with non-Kafka default transport
- **WHEN** a service starts in a profile where Kafka transport is not configured as default
- **THEN** producer and consumer fallback paths use shared noop transport implementations

#### Scenario: Consistent noop behavior across services
- **WHEN** multiple services run with noop transport fallback
- **THEN** behavior and interfaces are consistent because they use shared noop transport components

