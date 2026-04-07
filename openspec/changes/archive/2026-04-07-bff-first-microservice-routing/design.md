## Context

The platform already enforces api-gateway as the public ingress and validates identity before forwarding to api-service for orchestration. However, many non-auth frontend server routes still call jam/playback/catalog services directly. This creates split routing behavior and makes policy, observability, and contract evolution harder to enforce consistently.

## Goals / Non-Goals

**Goals:**
- Require all non-auth microservice HTTP calls from frontend server routes to traverse api-service (BFF) first.
- Keep auth lifecycle and explicit authN/authZ flows intact.
- Route realtime bootstrap config through BFF while keeping websocket transport direct to rt-gateway.
- Preserve normalized envelope behavior for frontend API responses.

**Non-Goals:**
- Change websocket transport to tunnel through BFF.
- Redesign core jam/playback/catalog domain semantics.
- Remove api-gateway from ingress role.

## Decisions

1. Mandatory non-auth HTTP route shape
- Decision: frontend route -> api-gateway -> api-service (BFF) -> downstream service.
- Rationale: one policy and contract control point for all non-auth microservice HTTP calls.

2. Keep auth boundary explicit
- Decision: auth lifecycle/auth validation remain dedicated auth flows; non-auth business calls move behind BFF.
- Rationale: avoids mixing identity lifecycle semantics with business aggregation/command flows.

3. Realtime bootstrap through BFF
- Decision: `/api/realtime/ws-config` is resolved through BFF; websocket connection remains direct to rt-gateway.
- Rationale: bootstrap is HTTP control-plane metadata; realtime data-plane remains ws direct for latency and fanout behavior.

4. Progressive rollout
- Decision: update sequence/mapping docs first (review gate), then implement route rewiring and BFF handlers.
- Rationale: reduces ambiguity and enables review-first alignment before code-wide refactor.

## Risks / Trade-offs

- [Risk] BFF throughput bottleneck due to expanded routing surface. -> Mitigation: endpoint-level timeout budgets, load testing, and staged rollout.
- [Risk] Temporary mismatch between docs and implementation during transition. -> Mitigation: keep explicit review note and complete route mapping updates in same change.
- [Risk] Contract drift across frontend and BFF adapters. -> Mitigation: enforce typed envelopes and focused route/integration tests.

## Migration Plan

1. Finalize and approve sequence + mapping contract.
2. Add/extend BFF route handlers for jam/playback/catalog/realtime bootstrap delegation.
3. Rewire frontend API routes to call BFF endpoints via api-gateway.
4. Run backend/frontend integration tests for command/query/realtime bootstrap paths.
5. Update runbook and verification flows.

## Open Questions

- Should `/api/auth/validate` remain direct auth flow or also be represented as a BFF-proxied auth context endpoint for consistency?
- For BFF route taxonomy, should command and query families be separated by explicit prefixes (for example `/v1/bff/mvp/commands/*` vs `/v1/bff/mvp/queries/*`)?
