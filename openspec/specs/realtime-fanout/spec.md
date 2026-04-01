# realtime-fanout Specification

## Purpose
Defines realtime websocket fanout behavior in `rt-gateway` for jam sessions, including ordered event delivery, gap recovery, reconnect fallback, and observability targets.
## Requirements
### Requirement: Session-scoped websocket room subscription
The `rt-gateway` service SHALL provide websocket room subscription for each jam session using room key `jam:{sessionId}`.

#### Scenario: Client subscribes to an active session room
- **WHEN** a client establishes websocket connection and subscribes with a valid `sessionId`
- **THEN** `rt-gateway` adds the client to room `jam:{sessionId}` and acknowledges subscription success

#### Scenario: Subscription rejected for invalid session identifier
- **WHEN** a client attempts to subscribe with missing or malformed `sessionId`
- **THEN** `rt-gateway` rejects the subscription request with a deterministic validation error

### Requirement: Kafka fanout consumer broadcasts queue and playback events
The `rt-gateway` service SHALL consume jam queue and playback event streams using consumer group `rt-gateway-fanout` and SHALL broadcast normalized updates to subscribers in room `jam:{sessionId}`.

#### Scenario: Queue event is consumed and broadcast
- **WHEN** `rt-gateway-fanout` consumes a valid queue event for a `sessionId`
- **THEN** the event is normalized and broadcast to all active subscribers in `jam:{sessionId}`

#### Scenario: Playback event is consumed and broadcast
- **WHEN** `rt-gateway-fanout` consumes a valid playback event for a `sessionId`
- **THEN** the event is normalized and broadcast to all active subscribers in `jam:{sessionId}`

### Requirement: Ordered delivery for a session stream
For a given `sessionId`, `rt-gateway` SHALL enforce monotonic event delivery by `aggregateVersion` and SHALL not broadcast out-of-order updates.

#### Scenario: Next sequential version arrives
- **WHEN** an event arrives with `aggregateVersion` equal to `lastBroadcastVersion + 1`
- **THEN** `rt-gateway` broadcasts the event and advances `lastBroadcastVersion`

#### Scenario: Duplicate or stale version arrives
- **WHEN** an event arrives with `aggregateVersion` less than or equal to `lastBroadcastVersion`
- **THEN** `rt-gateway` does not rebroadcast the event and records duplicate/stale telemetry

### Requirement: Gap detection triggers snapshot recovery
`rt-gateway` SHALL detect version gaps and SHALL recover session state by fetching authoritative snapshot from `jam-service` before resuming realtime fanout.

#### Scenario: Version gap detected
- **WHEN** an event arrives with `aggregateVersion` greater than `lastBroadcastVersion + 1`
- **THEN** `rt-gateway` pauses fanout for that `sessionId`, fetches session snapshot, broadcasts snapshot state, and resumes fanout from recovered version

#### Scenario: Snapshot fetch fails during recovery
- **WHEN** gap recovery snapshot fetch fails
- **THEN** `rt-gateway` emits recovery failure telemetry and retries according to configured backoff policy before resuming broadcast

### Requirement: Reconnect path supports cursor and snapshot fallback
`rt-gateway` SHALL support reconnect behavior that uses client cursor/version when available and SHALL provide snapshot fallback when cursor is behind current room version.

#### Scenario: Reconnect with current cursor
- **WHEN** a reconnecting client provides `lastSeenVersion` equal to current room version
- **THEN** `rt-gateway` resumes incremental fanout without snapshot replay

#### Scenario: Reconnect with stale cursor
- **WHEN** a reconnecting client provides `lastSeenVersion` lower than current room version with unrecoverable gap
- **THEN** `rt-gateway` sends authoritative snapshot before resuming incremental updates

### Requirement: Fanout latency and reliability are observable
The system SHALL expose fanout telemetry required to validate realtime performance and recovery behavior, including p95 fanout latency.

#### Scenario: Load test validates p95 latency
- **WHEN** load testing runs against realtime fanout flow for representative concurrent sessions
- **THEN** reported p95 fanout latency meets configured target and test evidence is recorded

#### Scenario: Recovery metrics are emitted
- **WHEN** gap detection and snapshot recovery occur
- **THEN** metrics for gap count, snapshot latency, and recovery outcome are emitted for operational monitoring

### Requirement: Websocket handshake MUST enforce configured origin allowlist
rt-gateway SHALL validate websocket handshake Origin against configured allowlist and SHALL reject unknown origins by default.

#### Scenario: Allowlisted origin succeeds
- **WHEN** a websocket client connects with Origin that exists in configured allowlist
- **THEN** handshake succeeds and connection is accepted

#### Scenario: Unknown origin is rejected
- **WHEN** a websocket client connects with Origin that is not allowlisted
- **THEN** handshake is rejected with deterministic forbidden-origin behavior and connection is closed

