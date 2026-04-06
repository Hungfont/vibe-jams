## Why

Jam rooms currently lack host moderation controls, so disruptive participants cannot be muted or removed with deterministic state propagation. ES-P2-002 is needed now to keep room control usable while preserving auditability and realtime consistency.

## What Changes

- Add jam-service moderation command handlers for host actions `mute` and `kick`.
- Persist moderation state in session participant projection and enforce blocked actions for muted or kicked users.
- Emit moderation audit events with `actor`, `target`, `reason`, and `timestamp` to Kafka topic `jam.moderation.events`.
- Extend rt-gateway consumer fanout to consume moderation topic and broadcast room updates.
- Add consumer hook interface in rt-gateway for abuse heuristic integrations.
- Add frontend moderation API routes and Jam Room participant controls so moderation actions are visible in UI updates.
- Update flow docs and runbook test flow to match moderation behavior.

## Capabilities

### New Capabilities
- `jam-moderation-controls`: Host moderation commands, participant moderation state transitions, blocked-action enforcement, and audit event payload contract.

### Modified Capabilities
- `kafka-event-foundation`: Topic baseline and ACL policy now includes moderation event stream.
- `realtime-fanout`: Fanout consumer now includes moderation events for room subscribers.

## Impact

- Affected code: `backend/jam-service/**`, `backend/rt-gateway/**`, `backend/shared/kafka/**`, `frontend/src/app/api/jam/**`, `frontend/src/components/jam/**`, `frontend/src/lib/jam/**`.
- Affected APIs: new jam-service moderation endpoints and frontend moderation route handlers.
- Affected Kafka topology: new topic `jam.moderation.events` and consumer subscription updates.
- Affected docs: `docs/frontend-backend-sequence.md`, `docs/runbooks/run.md`.
- Dependencies: participant registry, centralized authZ behavior from ES-P2-004, and rt-gateway consumer runtime.