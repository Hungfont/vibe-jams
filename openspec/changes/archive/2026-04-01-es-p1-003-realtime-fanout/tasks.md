## 1. Gateway Subscription Foundation

- [x] 1.1 Add websocket room subscription API in `rt-gateway` for `jam:{sessionId}` with sessionId validation
- [x] 1.2 Implement room membership lifecycle hooks (subscribe, unsubscribe, disconnect cleanup)
- [x] 1.3 Add connection context support for optional `lastSeenVersion` cursor on reconnect

## 2. Kafka Fanout Pipeline

- [x] 2.1 Add Kafka consumer group `rt-gateway-fanout` configuration and bootstrap wiring
- [x] 2.2 Consume queue and playback topics and map records into a normalized outbound websocket event envelope
- [x] 2.3 Implement per-session ordering guard using `aggregateVersion` with duplicate/stale suppression
- [x] 2.4 Add bounded fanout buffer policy and slow-consumer handling for websocket broadcasts

## 3. Gap Recovery and Snapshot Flow

- [x] 3.1 Implement version-gap detection (`incomingVersion > lastBroadcastVersion + 1`) in fanout path
- [x] 3.2 Add/consume `jam-service` snapshot read endpoint contract for authoritative session state
- [x] 3.3 Implement per-session recovery lock, snapshot fetch, snapshot broadcast, and stream resume logic
- [x] 3.4 Add retry with backoff and failure telemetry for snapshot recovery failures
- [x] 3.5 Implement reconnect decision flow to send snapshot fallback when client cursor is stale

## 4. Validation, Observability, and Rollout

- [x] 4.1 Add metrics for fanout latency, consumer lag, gap count, snapshot latency, and recovery outcomes
- [x] 4.2 Add integration tests for subscription, broadcast, ordering, duplicate suppression, and gap recovery scenarios
- [x] 4.3 Add reconnect tests covering current cursor and stale cursor snapshot fallback behavior
- [x] 4.4 Create load test scenario and verify p95 fanout latency target with recorded evidence
- [x] 4.5 Define feature-flag rollout and rollback checklist for staged deployment
