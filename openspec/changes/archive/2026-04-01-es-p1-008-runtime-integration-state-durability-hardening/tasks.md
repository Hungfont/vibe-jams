## 1. Runtime Configuration Baseline

- [x] 1.1 Define non-test adapter policy config for Kafka, durable repositories, catalog integration, auth validation, and websocket origin allowlist
- [x] 1.2 Define deterministic localhost fallback defaults for missing service-specific connections in local profile
- [x] 1.3 Add startup validation and fail-fast errors for invalid non-test adapter selections

## 2. Kafka Runtime Integration Cutover

- [x] 2.1 Replace mock or noop producer wiring with real Kafka integration in jam-service and playback-service non-test profiles
- [x] 2.2 Replace mock or noop fanout consumer wiring with real Kafka consumer integration in rt-gateway non-test profiles
- [x] 2.3 Add tests for adapter policy enforcement that reject noop transport in non-test profiles

## 3. Durable State Adapter Enforcement

- [x] 3.1 Switch jam session and queue runtime paths to durable adapters in non-test profiles
- [x] 3.2 Switch playback runtime state and sequencing metadata path to durable adapters in non-test profiles
- [x] 3.3 Add restart recovery tests for session, queue, and playback durability guarantees

## 4. Catalog and Auth Runtime Dependency Hardening

- [x] 4.1 Enforce runtime catalog validation integration and disallow in-memory-only catalog source in non-test profiles
- [x] 4.2 Enforce runtime token validation integration and disallow static in-memory claim fixtures in non-test profiles
- [x] 4.3 Add deterministic dependency-unavailable behavior tests for catalog and auth integration failures

## 5. Websocket Origin Security Hardening

- [x] 5.1 Implement origin allowlist check in rt-gateway handshake path
- [x] 5.2 Reject unknown origins by default with deterministic forbidden behavior and telemetry
- [x] 5.3 Add integration tests for allowlist success and unknown-origin rejection paths

## 6. Verification and Rollout Readiness

- [x] 6.1 Execute end-to-end runtime path test for queue or playback event fanout without mock adapters
- [x] 6.2 Execute concurrent command consistency tests under durable adapters and record evidence
- [x] 6.3 Update runtime setup and verification runbook with localhost fallback, dependency checks, and hardening test flow

## 7. Shared Noop Transport Standardization

- [x] 7.1 Add shared `NoOpsProducer` and `NoOpsConsumer` in `backend/shared` for Kafka fallback paths
- [x] 7.2 Replace service-local noop producer or consumer implementations with shared noop components in services that are not configured with Kafka as default transport
- [x] 7.3 Add or update tests to verify shared noop transport fallback behavior and compatibility with existing runtime adapter policies
