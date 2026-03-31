# Kafka Foundation Rollout Checklist (ES-P1-004)

## Provisioning Order

1. Create Phase 1 topics:
   - `jam.session.events` (12 partitions, 7d retention)
   - `jam.queue.events` (24 partitions, 7d retention)
   - `jam.playback.events` (12 partitions, 7d retention)
   - `analytics.user.actions` (24 partitions, 14d retention)
2. Apply producer ACLs:
   - `svc-jam-service` -> write `jam.session.events`, `jam.queue.events`
   - `svc-playback-service` -> write `jam.playback.events`
   - `svc-api-service` -> write `analytics.user.actions`
3. Apply consumer ACLs:
   - `svc-rt-gateway` -> read `jam.queue.events`, `jam.playback.events`

## Smoke Checks

- Publish one synthetic event per producer service and verify topic/key correctness.
- Run `backend/shared/cmd/topiccheck` with environment dump to verify partition/retention/ACL baseline.
- Validate that shared envelope parser can decode emitted events in consumer test pipeline.

## Rollback Switches

- Disable producer publish toggles in `jam-service`, `playback-service`, and `api-service`.
- Keep topic infrastructure in place to avoid recreate delays on re-enable.
- Re-run smoke checks before re-enabling publishers.
