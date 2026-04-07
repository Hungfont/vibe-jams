## MODIFIED Requirements

### Requirement: Kafka fanout consumer broadcasts queue and playback events
The `rt-gateway` service SHALL consume jam queue, playback, moderation, and permission event streams using consumer group `rt-gateway-fanout` and SHALL broadcast normalized updates to subscribers in room `jam:{sessionId}`.

#### Scenario: Queue event is consumed and broadcast
- **WHEN** `rt-gateway-fanout` consumes a valid queue event for a `sessionId`
- **THEN** the event is normalized and broadcast to all active subscribers in `jam:{sessionId}`

#### Scenario: Playback event is consumed and broadcast
- **WHEN** `rt-gateway-fanout` consumes a valid playback event for a `sessionId`
- **THEN** the event is normalized and broadcast to all active subscribers in `jam:{sessionId}`

#### Scenario: Moderation event is consumed and broadcast
- **WHEN** `rt-gateway-fanout` consumes a valid moderation event for a `sessionId`
- **THEN** the event is normalized and broadcast to all active subscribers in `jam:{sessionId}`

#### Scenario: Permission event is consumed and broadcast
- **WHEN** `rt-gateway-fanout` consumes a valid permission event for a `sessionId`
- **THEN** the event is normalized and broadcast to all active subscribers in `jam:{sessionId}`