## Why

Realtime Jam participants need low-latency, ordered session updates to stay synchronized while tracks and playback state change. Today, queue and playback events are produced, but there is no defined realtime fanout path in `rt-gateway` with gap recovery, which creates drift risk during reconnects and network loss.

## What Changes

- Add websocket room subscription flow in `rt-gateway` using room key `jam:{sessionId}`.
- Add Kafka consumer group `rt-gateway-fanout` to consume queue and playback events and broadcast them to room subscribers.
- Add ordered fanout guarantees for updates within the same `sessionId` stream.
- Add reconnect behavior with version-gap detection and snapshot-recovery fetch from `jam-service`.
- Add load/performance validation for fanout latency with p95 target verification.

## Capabilities

### New Capabilities
- `realtime-fanout`: Session-scoped websocket fanout from Kafka events with ordered delivery and snapshot recovery for version gaps.

### Modified Capabilities
- None.

## Impact

- Affected services: `rt-gateway`, `jam-service`.
- Affected runtime dependencies: Kafka topics for jam queue/playback events, consumer group `rt-gateway-fanout`, websocket infrastructure.
- Potential API impact: `jam-service` session state read endpoint may require stronger snapshot/read contract for recovery.
- Operational impact: Add fanout load test and observability around lag, ordering, and recovery path frequency.
