# Engineer Stories - Spotify-Like Jam

## Phase 1 (MVP)

### ES-P1-001 - Jam Session APIs
- Task Description:
  - Implement session lifecycle endpoints in `jam-service`: create, join, leave, end.
  - Enforce premium check for host creation and host ownership for end.
  - Persist session metadata and permissions in Redis keyspace.
- Dependencies:
  - Auth token validation and plan claim in `auth-service`.
  - Redis connectivity and key schema.
- Definition of Done:
  - Endpoints return expected status codes and payloads.
  - Host leave ends session and blocks further writes.
  - Unit and API tests for success + error paths pass.

### ES-P1-002 - Queue Command Handling
- Task Description:
  - Implement add/remove/reorder queue operations with atomic writes and `queueVersion` increment.
  - Add idempotency handling for queue-add command.
  - Return `409 version_conflict` on stale reorder requests.
- Dependencies:
  - Redis list/hash primitives and version key.
  - Shared error model in API layer.
- Definition of Done:
  - Concurrent write tests show deterministic queue state.
  - Duplicate idempotency key does not duplicate queue item.
  - Queue snapshot reflects latest version after each mutation.

### ES-P1-003 - Realtime Fanout
- Task Description:
  - Build websocket room subscription flow for `jam:{sessionId}` in `rt-gateway`.
  - Consume Kafka queue/playback events and broadcast to room subscribers.
  - Implement reconnect and snapshot-recovery behavior for version gaps.
- Dependencies:
  - Kafka consumer group `rt-gateway-fanout`.
  - Session state read endpoint in `jam-service`.
- Definition of Done:
  - Clients receive ordered updates for same `sessionId`.
  - Gap detection triggers snapshot fetch path.
  - Load test validates p95 fanout latency target.

### ES-P1-004 - Kafka Foundation
- Task Description:
  - Create Kafka topics for session, queue, playback, analytics events.
  - Implement producer libraries in `jam-service`, `playback-service`, `api-service`.
  - Standardize event envelope with `eventId`, `sessionId`, `aggregateVersion`.
- Dependencies:
  - Kafka cluster provisioning and ACLs.
  - Shared schema package (JSON contract).
- Definition of Done:
  - All producers publish with expected keying (`sessionId`/`userId`).
  - Consumer can parse and validate envelope consistently.
  - Retention and partition settings match LLD.

### ES-P1-005 - Playback Command Pipeline
- Task Description:
  - Implement playback command endpoint and internal command executor.
  - Publish playback state transitions to Kafka.
  - Synchronize playback transitions with queue version checks.
- Dependencies:
  - Queue state read from Redis.
  - Kafka topic `jam.playback.events`.
- Definition of Done:
  - Host commands accepted and emitted to Kafka.
  - Unauthorized/stale commands return correct errors.
  - Integration tests validate command-to-event path.

### ES-P1-006 - Auth Entitlement Guard for Jam
- Task Description:
  - Implement token validation and premium entitlement checks consumed by Jam create/end flows.
  - Provide stable auth claim contract for user identity and plan.
  - Standardize `401 unauthorized` and `403 premium_required` mapping.
- Dependencies:
  - Auth token/session store integration.
  - Shared auth middleware contract in API/Jam layer.
- Definition of Done:
  - Jam session create/end authorization is no longer blocked by auth integration gaps.
  - Premium/non-premium behavior matches policy in automated tests.
  - Error responses are consistent across entrypoints.

### ES-P1-007 - Catalog Track Validation API
- Task Description:
  - Implement catalog lookup for `trackId` with playable metadata contract.
  - Add track existence/availability validation for queue and playback pre-checks.
  - Return deterministic error for missing/unavailable tracks.
- Dependencies:
  - Track metadata source and storage contract.
  - Integration points in `jam-service` and `playback-service`.
- Definition of Done:
  - Queue and playback commands can validate track data without ad-hoc logic.
  - Invalid tracks are rejected before state mutation.
  - Contract tests verify catalog response schema.

## Phase 2 (Jam Control and Consistency Upgrades)

### ES-P2-001 - Permission Projection
- Task Description:
  - Implement granular guest permissions (`canControlPlayback`, `canReorderQueue`, `canChangeVolume`).
  - Persist and publish permission updates via `jam.permission.events`.
  - Enforce permission checks at command entrypoint.
- Dependencies:
  - Host authZ guard.
  - Kafka topic `jam.permission.events`.
- Definition of Done:
  - Permission toggle reflects in realtime for active sessions.
  - Restricted actions blocked immediately for guests.
  - Permission changes are auditable by event stream.

### ES-P2-002 - Moderation Controls
- Task Description:
  - Implement host moderation actions: mute and kick.
  - Broadcast moderation events and update participant state.
  - Add consumer hooks for abuse heuristics.
- Dependencies:
  - Participant registry in session state.
  - Kafka topic `jam.moderation.events`.
- Definition of Done:
  - Muted/kicked user cannot perform blocked actions.
  - Moderation actions are visible in room updates.
  - Audit trail contains actor, target, reason, timestamp.

### ES-P2-003 - Stronger Concurrency Guards
- Task Description:
  - Add mandatory `expectedQueueVersion` for reorder/remove writes.
  - Introduce retry guidance payload in `409` responses.
  - Add `playbackEpoch` to playback state update contract.
- Dependencies:
  - Shared response schema for conflict errors.
  - Client-side version reconciliation support.
- Definition of Done:
  - Stale writes deterministically rejected.
  - Playback updates include both epoch and queue version.
  - Contract tests verify schema compatibility.

### ES-P2-004 - AuthZ Guard for Permission and Moderation
- Task Description:
  - Implement centralized host/guest authorization guard for permission and moderation commands.
  - Attach actor identity context to audit events for policy actions.
  - Ensure host-only actions fail fast with consistent `403 host_only`.
- Dependencies:
  - Auth claim contract from auth-service.
  - Audit event schema for moderation/permission actions.
- Definition of Done:
  - Non-host actors cannot perform host-only policy actions.
  - Authorization behavior is consistent across Jam command entrypoints.
  - Audit logs contain actor metadata for denied and accepted actions.

### ES-P2-005 - Catalog Policy Checks for Queue Operations
- Task Description:
  - Add catalog-level policy checks for queue commands (availability/restriction).
  - Expose policy status needed by Jam queue handlers.
  - Keep compatibility with existing idempotency and version checks.
- Dependencies:
  - Queue command integration points in `jam-service`.
  - Catalog policy rules source.
- Definition of Done:
  - Restricted/unavailable tracks are rejected with deterministic errors.
  - Queue consistency logic remains unchanged by policy check rollout.
  - Regression tests cover policy-on and policy-off cases.

## Phase 3 (Personalization and Recommendation)

### ES-P3-001 - Feature Event Pipeline
- Task Description:
  - Emit recommendation-relevant interactions (skip, add, like) to `recommendation.features`.
  - Build feature consumer to aggregate short-window session features.
  - Store feature snapshots for online ranking.
- Dependencies:
  - Existing Jam and playback event streams.
  - Kafka topic `recommendation.features`.
- Definition of Done:
  - Feature events are emitted with session/user context.
  - Aggregator computes and stores rolling feature vectors.
  - Monitoring confirms no significant consumer lag.

### ES-P3-002 - Online Ranking and Serving
- Task Description:
  - Implement recommendation ranker consumer and serving endpoint.
  - Publish result updates to `recommendation.outputs`.
  - Integrate recommendation add-to-queue path with existing idempotency/version checks.
- Dependencies:
  - Feature builder output.
  - Kafka topic `recommendation.outputs`.
- Definition of Done:
  - Recommendation API returns ranked list per active session.
  - Suggested track can be added to queue in one step.
  - Ranking updates are traceable by session.

### ES-P3-003 - Recommendation Access Scope Validation
- Task Description:
  - Add auth scope validation for recommendation serving endpoints.
  - Validate user/session access before returning recommendation payload.
  - Standardize unauthorized access errors without leaking session internals.
- Dependencies:
  - Auth claim verification flow.
  - Session context validation contract.
- Definition of Done:
  - Unauthorized users cannot access recommendation payloads.
  - Authorized requests keep expected latency profile.
  - Endpoint tests cover scope mismatch and valid access paths.

### ES-P3-004 - Recommendation Catalog Enrichment
- Task Description:
  - Enrich recommendation outputs with required catalog track metadata.
  - Add fallback/filter behavior for missing catalog records.
  - Ensure enriched payload supports one-step add-to-queue.
- Dependencies:
  - Catalog lookup API for recommendation results.
  - Recommendation serving contract in API layer.
- Definition of Done:
  - Recommendation responses include renderable track data.
  - Missing metadata does not break serving endpoint.
  - Add-to-queue from recommendation remains functional end-to-end.

## Phase 4 (Hardening and Scale)

### ES-P4-001 - Kafka Operational Hardening
- Task Description:
  - Rebalance partitions based on observed throughput.
  - Enable stricter producer acks and idempotent producers where required.
  - Formalize retention and optional compaction policy by topic class.
- Dependencies:
  - Throughput/lag baseline from production metrics.
  - Platform SRE change window.
- Definition of Done:
  - No data-loss regression in failure simulation.
  - Consumer lag remains within agreed threshold post-change.
  - Topic policy matrix documented and approved.

### ES-P4-002 - Realtime Backpressure and Resilience
- Task Description:
  - Implement backpressure controls in websocket fanout pipeline.
  - Add circuit-breaker behavior for downstream dependency failures.
  - Add graceful degradation mode for non-critical updates.
- Dependencies:
  - Realtime gateway instrumentation hooks.
  - Failover policy and runbook.
- Definition of Done:
  - Under stress, critical playback events are prioritized.
  - Gateway remains healthy under burst traffic conditions.
  - Recovery behavior verified in fault-injection test.

### ES-P4-003 - Multi-AZ Recovery Drills
- Task Description:
  - Define and execute failover drills for core services and Kafka consumers.
  - Validate session continuity and data consistency after failover.
  - Document recovery SOP and rollback criteria.
- Dependencies:
  - Multi-AZ deployment support.
  - Operational runbook ownership.
- Definition of Done:
  - Drill results include RTO/RPO measurements.
  - No unrecoverable session/data corruption in simulation.
  - SOP approved by engineering and operations stakeholders.

### ES-P4-004 - Identity-Aware Abuse and Rate Controls
- Task Description:
  - Implement identity-aware rate limit keys using auth claims.
  - Add abuse decision hook integration for temporary identity/session blocks.
  - Ensure controls prioritize protection without broad false-positive impact.
- Dependencies:
  - Auth claim propagation across API and gateway boundaries.
  - Policy/rate control configuration source.
- Definition of Done:
  - Abuse controls can throttle/block by identity context.
  - Legitimate session traffic remains within target SLO under policy load.
  - Security tests verify block and unblock workflows.

### ES-P4-005 - Catalog Resilience and Fallback
- Task Description:
  - Implement degraded-mode read fallback for catalog dependency failures.
  - Define recovery transition logic to return to normal consistency mode.
  - Add runbook validation steps for catalog failure and recovery drills.
- Dependencies:
  - Catalog data source redundancy strategy.
  - Operational incident runbook ownership.
- Definition of Done:
  - Playback/recommendation flows remain partially available during catalog incidents.
  - Recovery to healthy mode is deterministic and tested.
  - Fault-injection tests verify fallback behavior under stress.
