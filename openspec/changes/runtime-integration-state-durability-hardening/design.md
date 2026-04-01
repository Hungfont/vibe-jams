## Context

Phase-1/Phase-3 capabilities established queue, playback, auth, catalog validation, and realtime fanout contracts, but runtime paths still permit mock/noop adapters in several bootstraps. This creates production-readiness gaps:

- Runtime event propagation can silently degrade when Kafka wiring is not active.
- Session/queue/playback state can be lost on restart when in-memory adapters are selected.
- Catalog and auth validation can diverge when non-durable local-only providers are used in runtime mode.
- Websocket handshake policy is not strict enough for hardened environments.

This change is cross-service and spans `jam-service`, `playback-service`, `rt-gateway`, and integration boundaries with `catalog-service` and `auth-service`.

## Goals / Non-Goals

**Goals:**
- Require real Kafka integration for runtime queue/playback/fanout transport outside explicit test mode.
- Require durable persistence adapters (Redis/PostgreSQL as applicable) for session, queue, playback, catalog source, and auth validation flows.
- Enforce localhost fallback defaults for local development when explicit endpoints are not provided.
- Enforce websocket origin allowlist at handshake with reject-by-default behavior.
- Add verification coverage for restart durability, concurrent command consistency, and origin allowlist behavior.

**Non-Goals:**
- Re-architecting event schemas or introducing new business command types.
- Changing entitlement business policy beyond validator source hardening.
- Building a full infrastructure provisioning system for Kafka/Redis/PostgreSQL.

## Decisions

1. Runtime profile gating for adapters
- Introduce explicit runtime-mode checks so noop/in-memory adapters are permitted only in test profile, not in normal runtime profiles.
- Startup validation fails fast when required Kafka or persistence adapters cannot be initialized in runtime mode.
- Alternative considered: best-effort downgrade to in-memory when dependencies unavailable. Rejected because it masks operational failures and violates durability requirements.

2. Durable state adapters as default runtime path
- `jam-service` and `playback-service` must use durable repository implementations for command state and idempotency-relevant metadata.
- `catalog-service` lookup and auth validation integrations must read from real configured backends in runtime mode.
- Alternative considered: hybrid write-through with in-memory primary and durable async sink. Rejected due to consistency and restart recovery complexity.

3. Localhost-first fallback policy for local development
- If service-specific connection strings are absent in local profile, default to localhost conventions for Kafka and Redis.
- PostgreSQL uses configured DSN environment variable; documentation references secure secret injection and does not embed credentials in code.
- Alternative considered: hard fail on any missing connection config even in local. Rejected to preserve developer velocity.

4. Strict websocket origin allowlist enforcement
- `rt-gateway` handshake path requires configured allowlist matching; unknown origins are rejected.
- Empty allowlist defaults to deny-all in non-local runtime and explicit local dev origin list in local profile.
- Alternative considered: wildcard allow with warning telemetry. Rejected due to security hardening objective.

5. Verification strategy prioritizes behavior over unit internals
- Add integration tests for cross-service fanout path with real Kafka wiring.
- Add durability tests proving state survives restart and remains consistent under concurrent commands.
- Add handshake tests for websocket origin allow and reject flows.

## Risks / Trade-offs

- [Risk] Stricter startup validation can increase deployment failures during initial rollout. -> Mitigation: staged rollout with clear config validation logs and preflight checks.
- [Risk] Real backend dependencies can increase local setup complexity. -> Mitigation: deterministic localhost defaults and explicit runbook setup flows.
- [Risk] Durable adapters may increase write latency compared with in-memory mocks. -> Mitigation: optimize repository calls, pool connections, and monitor p95 command latency.
- [Risk] Origin allowlist misconfiguration can block valid clients. -> Mitigation: environment-specific allowlist templates and handshake rejection telemetry.

## Migration Plan

1. Add runtime configuration schema updates and startup validation guards for each service.
2. Wire Kafka producer/consumer integrations as default runtime path in `jam-service`, `playback-service`, and `rt-gateway`.
3. Replace in-memory repositories/adapters with durable Redis/PostgreSQL adapters in runtime mode.
4. Roll out websocket origin allowlist enforcement behind environment configuration gates.
5. Execute integration, restart durability, and websocket policy test flows; update runbooks for operational verification.

Rollback approach:
- Allow temporary feature gates to disable strict runtime enforcement per service while keeping code paths intact.
- Revert to previous deployment image if critical runtime integration errors occur, then fix configuration and re-roll forward.

## Open Questions

- Which non-local environments require mandatory PostgreSQL for each service versus optional integration mode?
- Should auth token validation cache TTLs be standardized across `jam-service` and `playback-service` in this phase?
- What is the minimum required websocket origin set for staging and production gateways?
