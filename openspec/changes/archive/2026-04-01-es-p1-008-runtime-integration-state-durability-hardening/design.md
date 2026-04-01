## Context

Current MVP behavior depends on a mix of real and mock runtime adapters. Queue, playback, and fanout paths can diverge from production behavior when non-durable or noop paths are active. ES-P1-008 hardens runtime integration by requiring real dependencies in non-test profiles and by enforcing state durability and websocket origin policy.

## Goals / Non-Goals

**Goals:**
- Enforce real Kafka integration for jam-service, playback-service, and rt-gateway fanout in non-test profiles.
- Enforce durable runtime state adapters for jam session, queue, and playback state.
- Ensure catalog lookup and auth token validation use real runtime integrations in non-test profiles.
- Enforce websocket origin allowlist with deny-by-default behavior.
- Provide deterministic local fallback configuration for missing explicit service connections.

**Non-Goals:**
- Redesigning domain command contracts for queue, playback, or sessions.
- Introducing new recommendation or moderation features.
- Replacing existing event envelope schema.

## Decisions

1. Runtime profile enforcement
- Introduce adapter policy checks at startup.
- Non-test profiles MUST fail fast if mock or in-memory adapters are selected for hardened paths.
- Test profile keeps mock adapters for isolated testing.

2. Durable adapter cutover
- Jam and playback command state uses durable repositories in runtime mode.
- Idempotency and version metadata must persist across restart boundaries.
- Local dev can use local Redis and local Postgres connection defaults when explicit configuration is missing.

3. Real integration for catalog and auth dependencies
- Runtime catalog validation uses configured catalog integration, not in-memory fixtures.
- Runtime auth claim validation uses configured token validation integration, with deterministic unauthorized and dependency-unavailable handling.

4. Websocket origin hardening
- Handshake checks Origin against configured allowlist.
- Unknown origins are rejected with deterministic forbidden behavior.
- Local profile allowlist defaults are explicit and environment-scoped.

5. Verification strategy
- Add integration tests for runtime Kafka fanout path.
- Add restart durability tests for session, queue, and playback state.
- Add websocket allow and reject handshake tests.

## Risks / Trade-offs

- [Risk] Stricter startup checks can block deployments with incomplete configuration. -> Mitigation: config preflight and clear startup diagnostics.
- [Risk] Durable storage can increase latency under load. -> Mitigation: connection pooling, bounded retries, and p95 monitoring.
- [Risk] Origin allowlist misconfiguration can block legitimate clients. -> Mitigation: environment templates and explicit rollout validation.

## Migration Plan

1. Add configuration schema and startup validation for adapter policy and origin allowlist.
2. Enable real Kafka runtime wiring for jam-service, playback-service, and rt-gateway.
3. Switch queue, session, and playback runtime repositories to durable adapters in non-test profiles.
4. Switch catalog and auth runtime integrations to real providers in non-test profiles.
5. Run integration and durability tests, then roll out progressively by environment.
6. Keep rollback switches to temporarily re-enable previous runtime profile behavior only during controlled rollback windows.

## Open Questions

- Which services require PostgreSQL in MVP runtime versus Redis-only paths?
- What are the exact allowed origins for staging and production gateway deployments?
- Should dependency-unavailable errors for catalog and auth be unified under one error contract in MVP?
