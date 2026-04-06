## MODIFIED Requirements

### Requirement: Kafka fanout consumer broadcasts queue and playback events
The `rt-gateway` service SHALL consume jam queue, playback, and moderation event streams using consumer group `rt-gateway-fanout` and SHALL broadcast normalized updates to subscribers in room `jam:{sessionId}`.

#### Scenario: Queue event is consumed and broadcast
- **WHEN** `rt-gateway-fanout` consumes a valid queue event for a `sessionId`
- **THEN** the event is normalized and broadcast to all active subscribers in `jam:{sessionId}`

#### Scenario: Playback event is consumed and broadcast
- **WHEN** `rt-gateway-fanout` consumes a valid playback event for a `sessionId`
- **THEN** the event is normalized and broadcast to all active subscribers in `jam:{sessionId}`

#### Scenario: Moderation event is consumed and broadcast
- **WHEN** `rt-gateway-fanout` consumes a valid moderation event for a `sessionId`
- **THEN** the event is normalized and broadcast to all active subscribers in `jam:{sessionId}`

### Requirement: Fanout latency and reliability are observable
The system SHALL expose fanout telemetry required to validate realtime performance and recovery behavior, including p95 fanout latency.

#### Scenario: Load test validates p95 latency
- **WHEN** load testing runs against realtime fanout flow for representative concurrent sessions
- **THEN** reported p95 fanout latency meets configured target and test evidence is recorded

#### Scenario: Recovery metrics are emitted
- **WHEN** gap detection and snapshot recovery occur
- **THEN** metrics for gap count, snapshot latency, and recovery outcome are emitted for operational monitoring

#### Scenario: Moderation consumer hook can observe moderation envelopes
- **WHEN** moderation envelope is consumed in fanout consumer loop
- **THEN** configured abuse-heuristic hook is invoked with envelope metadata for downstream policy analysis