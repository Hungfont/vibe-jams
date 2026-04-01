## Why

Current runtime flows for jam queue, playback, and realtime fanout still allow mock or in-memory adapters in places where production behavior requires durable storage and real event transport. This creates restart data loss risk, drift between services, and weak security defaults for websocket origin checks.

## What Changes

- Replace mock or noop Kafka runtime wiring with real producer/consumer integration across `jam-service`, `playback-service`, and `rt-gateway` fanout paths.
- Replace in-memory state adapters with durable Redis/PostgreSQL-backed adapters for jam session and queue state, playback state, catalog lookup source, and auth token validation support paths.
- Add deterministic runtime configuration fallback rules so local development defaults to localhost endpoints when service-specific connections are not provided.
- Enforce websocket origin allowlist policy in realtime gateway handshake and reject unknown origins by default.
- Add integration and recovery-focused test coverage for restart durability, concurrent command consistency, and websocket origin allowlist behavior.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `kafka-event-foundation`: Runtime requirements updated to require real broker wiring for queue/playback/fanout paths and to disallow noop transport in non-test runtime modes.
- `jam-session-lifecycle`: Session persistence requirements updated to require restart-safe Redis/PostgreSQL-backed storage paths and consistency under concurrent lifecycle mutations.
- `jams-queue-command-handling`: Queue command handling requirements updated to require durable state writes and deterministic recovery after service restart.
- `jam-playback-command-pipeline`: Playback command pipeline requirements updated to require durable playback state and idempotent replay-safe behavior under restart conditions.
- `catalog-track-validation-api`: Catalog validation requirements updated to require durable catalog source integration for lookup behavior (no in-memory-only source in runtime mode).
- `auth-claim-contract`: Claim validation requirements updated to require runtime integration with real token validation source and deterministic failure behavior when external validation is unavailable.
- `realtime-fanout`: Realtime gateway requirements updated to enforce websocket origin allowlist policy and reject unknown origins by default.

## Impact

- Affected services: `jam-service`, `playback-service`, `rt-gateway`, `catalog-service`, `auth-service`, and shared Kafka/auth/catalog contracts.
- Affected runtime dependencies: Kafka, Redis, PostgreSQL, and environment-based connection configuration for local versus non-local environments.
- Security impact: websocket handshake policy becomes deny-by-default for non-allowlisted origins.
- Reliability impact: command and fanout flows must continue with durable state semantics across process restarts and concurrent command load.
- Testing impact: integration, recovery, and websocket-origin policy tests are required to satisfy definition of done.
