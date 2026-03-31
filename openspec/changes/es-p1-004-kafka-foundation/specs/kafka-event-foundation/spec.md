## ADDED Requirements

ALWAYS implement code as microservice rules.

### Requirement: Phase 1 Kafka topics are provisioned with defined policies
The system SHALL provision `jam.session.events`, `jam.queue.events`, `jam.playback.events`, and `analytics.user.actions` with partition and retention settings aligned to Phase 1 LLD.

#### Scenario: Topic provisioning matches expected settings
- **WHEN** infrastructure provisioning completes for an environment
- **THEN** each required topic exists and its partition and retention configuration matches the defined Phase 1 values

### Requirement: Producers publish with deterministic topic keying
The system SHALL publish jam session, queue, and playback events using `sessionId` as the Kafka key and SHALL publish analytics events using `userId` as the Kafka key.

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
