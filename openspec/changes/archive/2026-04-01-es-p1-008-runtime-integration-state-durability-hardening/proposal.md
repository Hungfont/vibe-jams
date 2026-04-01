## Why

Phase 1 runtime flows still allow mock and in-memory wiring in paths that must behave like production. This causes inconsistent cross-service behavior, restart data loss risk, and weak websocket boundary controls.

## What Changes

- Replace mock runtime Kafka wiring with real producer and consumer integration for jam queue, playback, and realtime fanout paths.
- Replace in-memory runtime adapters with durable Redis and PostgreSQL adapters for session, queue, and playback state.
- Ensure runtime catalog lookup and auth token validation use real dependency integrations in non-test profiles.
- Add deterministic local fallback configuration for services that do not provide explicit connection settings.
- Enforce websocket origin allowlist policy and reject unknown origins by default.
- Add durability and security verification coverage for restart recovery, concurrent commands, and websocket origin checks.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `kafka-event-foundation`: Runtime behavior updated to require real Kafka transport in non-test profiles for queue, playback, and fanout flows.
- `jams-queue-command-handling`: Queue command requirements updated to require durable runtime persistence and restart-safe idempotency behavior.
- `jam-playback-command-pipeline`: Playback command requirements updated to require durable runtime state and replay-safe transition handling.
- `jam-session-lifecycle`: Session lifecycle requirements updated to require durable state recovery after restart.
- `realtime-fanout`: Realtime gateway requirements updated to enforce websocket origin allowlist and deny unknown origins.
- `catalog-track-validation-api`: Catalog validation requirements updated to require real runtime lookup integration (no in-memory-only runtime source in non-test profiles).
- `auth-claim-contract`: Auth claim requirements updated to require runtime token validation integration and deterministic failure semantics.

## Impact

- Affected services: jam-service, playback-service, rt-gateway, catalog-service, auth-service.
- Affected dependencies: Kafka, Redis, PostgreSQL, auth validation provider, catalog source.
- Affected runtime configuration: connection defaults, non-test adapter restrictions, websocket allowed-origin settings.
- Test impact: integration and resilience tests for restart durability, concurrent command consistency, and websocket origin allowlist behavior.
