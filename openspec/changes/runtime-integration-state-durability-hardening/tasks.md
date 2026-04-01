## 1. Runtime Configuration and Guardrails

- [ ] 1.1 Define per-service runtime config schema for Kafka, Redis, PostgreSQL, auth validation, and websocket origin allowlist
- [ ] 1.2 Implement localhost fallback defaults for local profile when service-specific endpoints are not provided
- [ ] 1.3 Add startup validation that fails fast when runtime profile selects missing/invalid real adapters

## 2. Kafka Runtime Integration Hardening

- [ ] 2.1 Replace noop/mock producer wiring with real Kafka producer wiring in `jam-service` and `playback-service` runtime bootstraps
- [ ] 2.2 Replace noop/mock fanout consumer wiring with real Kafka consumer wiring in `rt-gateway`
- [ ] 2.3 Restrict noop/mock transport mode to explicit test profile and add validation tests for non-test rejection

## 3. Durable State Adapter Migration

- [ ] 3.1 Replace in-memory jam session and queue adapters with durable Redis/PostgreSQL-backed repositories in runtime profile
- [ ] 3.2 Replace in-memory playback state adapters with durable runtime repositories and persist sequencing/idempotency metadata
- [ ] 3.3 Replace in-memory catalog and auth validation runtime adapters with configured durable/external-backed integrations

## 4. Realtime Origin Security Enforcement

- [ ] 4.1 Enforce websocket handshake origin allowlist in `rt-gateway` with reject-by-default behavior for unknown origins
- [ ] 4.2 Add deterministic handshake error response and telemetry fields for rejected origins
- [ ] 4.3 Add allowlist and reject-path integration tests for websocket connection policy

## 5. Durability and Concurrency Verification

- [ ] 5.1 Add restart recovery integration tests proving queue/playback/session state survives service restarts
- [ ] 5.2 Add concurrent command consistency tests for queue/playback/session write paths under durable adapters
- [ ] 5.3 Add cross-service integration test covering queue mutation -> Kafka event -> realtime fanout path without mock transport

## 6. Operational Readiness and Documentation

- [ ] 6.1 Document runtime dependency requirements and secure environment variable setup (Kafka/Redis/PostgreSQL/auth)
- [ ] 6.2 Update `docs/runbooks/run.md` with concise test flows for runtime integration, restart durability, and websocket origin allowlist verification
- [ ] 6.3 Capture rollout and rollback checklist for strict runtime adapter enforcement across environments
