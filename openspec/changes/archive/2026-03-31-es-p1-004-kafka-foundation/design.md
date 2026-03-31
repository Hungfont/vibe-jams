## Context

Phase 1 introduces Kafka as the event backbone for jam session state changes and analytics actions. Multiple services publish events (`jam-service`, `playback-service`, `api-service`) and downstream consumers such as `rt-gateway` depend on consistent ordering, keying, and envelope parsing. The LLD already defines target topics, keying (`sessionId` for jam streams, `userId` for analytics), and baseline partition/retention settings.

## Goals / Non-Goals

**Goals:**
- Implement a shared event envelope contract that all producers serialize consistently.
- Ensure producer implementations publish to the correct Phase 1 topics with correct keys.
- Enforce LLD-aligned retention and partition settings at topic provisioning time.
- Provide validation and tests that protect against envelope/keying drift.

**Non-Goals:**
- Implementing consumer-side business workflows beyond parse/validation compatibility.
- Solving multi-region Kafka topology or disaster recovery in this phase.
- Introducing schema registry and Avro/Protobuf migration in MVP.

## Decisions

1. Shared contract-first envelope
   - Decision: define one shared JSON envelope contract and require all producers to emit that format.
   - Rationale: consumer compatibility and reduced producer divergence.
   - Alternative considered: per-service custom envelopes; rejected due to parsing complexity and higher integration risk.

2. Service-local producer adapters over shared base utilities
   - Decision: keep thin producer adapters in each service while sharing envelope model and validation helpers.
   - Rationale: balances consistency with service autonomy and simpler ownership boundaries.
   - Alternative considered: one centralized producer SDK that owns all topic routing; rejected because it increases coupling and rollout risk.

3. Topic/key strategy aligned with event semantics
   - Decision: use `sessionId` as key for jam session, queue, and playback topics; use `userId` for analytics topic.
   - Rationale: preserves per-session ordering where required while distributing analytics throughput.
   - Alternative considered: key all events by `sessionId`; rejected for analytics because user-centric partitioning better matches expected aggregations.

4. Provisioning as explicit infrastructure task
   - Decision: provision topics and ACLs explicitly per environment before producer rollout.
   - Rationale: avoids runtime create-topic races and permission failures in production paths.
   - Alternative considered: auto-create topics from services; rejected because it weakens governance for retention/partition policies.

5. Validation at producer boundary
   - Decision: validate required envelope fields and key presence before publish.
   - Rationale: fail fast at source and prevent bad events from entering streams.
   - Alternative considered: rely only on consumer-side validation; rejected because invalid events still pollute topics.

## Risks / Trade-offs

- [Config drift between LLD and deployed topics] -> Mitigation: codify topic definitions in versioned config and validate at startup/CI.
- [Envelope version evolution breaks consumers] -> Mitigation: start with additive-only changes and include contract tests for known consumers.
- [Hot partition risk on large sessions] -> Mitigation: monitor partition skew and revisit partition counts in Phase 4 hardening.
- [Operational failures from ACL misconfiguration] -> Mitigation: pre-deploy ACL checks and environment readiness validation.

## Migration Plan

1. Provision topics and ACLs in non-prod using LLD defaults.
2. Roll out shared envelope contract and producer adapters behind feature toggles.
3. Enable publishing from `jam-service`, then `playback-service`, then `api-service` with verification metrics.
4. Validate consumer parse compatibility in `rt-gateway` and analytics pipelines.
5. Promote to production gradually; rollback by disabling producer publish toggles while retaining topic infrastructure.

## Open Questions

- Should envelope include an explicit `schemaVersion` field in MVP, or defer to Phase 4 hardening?
- Do analytics events require stronger delivery guarantees (idempotent producer settings) in Phase 1?
- Is topic lifecycle managed by Terraform, script-based tooling, or platform team automation for this repository?
