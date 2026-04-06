## MODIFIED Requirements

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