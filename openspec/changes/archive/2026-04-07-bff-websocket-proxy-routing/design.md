## Context

The current implementation routes websocket bootstrap config through api-gateway -> api-service (BFF), but the resulting `wsUrl` still points clients directly to rt-gateway `/ws`. This creates split ingress behavior where realtime control-plane follows BFF-first, but realtime data-plane bypasses it at connect time.

## Goals / Non-Goals

**Goals:**
- Ensure frontend websocket connects use gateway/BFF path, not direct rt-gateway URL.
- Preserve current auth behavior at gateway middleware for protected routes, including cookie fallback.
- Preserve rt-gateway fanout semantics while changing ingress path only.
- Keep frontend contract deterministic (`wsUrl`, `sessionId`, `lastSeenVersion`) with updated proxy target.

**Non-Goals:**
- Redesign realtime fanout internals, Kafka processing, or snapshot recovery semantics.
- Introduce a new websocket protocol or payload schema.
- Remove existing bootstrap endpoint shape.

## Decisions

1. Add BFF websocket proxy endpoint
- Decision: introduce BFF websocket proxy path (gateway-facing), proxied by api-service to rt-gateway `/ws` with query passthrough.
- Rationale: satisfies BFF-first/no-direct-calling requirement with minimal client contract changes.

2. Keep gateway as public ingress for websocket connect
- Decision: browser websocket target becomes gateway-hosted BFF route instead of rt-gateway origin.
- Rationale: keeps single public ingress and reuses existing gateway auth middleware enforcement.

3. Bootstrap contract returns proxied wsUrl
- Decision: ws-config response returns websocket URL pointing at gateway/BFF websocket proxy path.
- Rationale: frontend client remains simple (single URL usage) while eliminating direct rt-gateway exposure.

4. Path rewrite in BFF websocket proxy
- Decision: BFF proxy rewrites gateway-facing websocket path to downstream rt-gateway `/ws` path while preserving query parameters (`sessionId`, `lastSeenVersion`).
- Rationale: avoids rt-gateway API changes and isolates routing adaptation in BFF layer.

## Risks / Trade-offs

- [Risk] Additional proxy hop can increase websocket handshake latency. -> Mitigation: keep lightweight reverse proxy path and add focused route tests.
- [Risk] Upgrade header/protocol handling regressions in proxy chain. -> Mitigation: test proxy route behavior for upgrade-like requests and verify end-to-end local flow.
- [Risk] Auth cookie scope mismatch could reject websocket connects. -> Mitigation: retain gateway cookie-fallback auth behavior and validate ws bootstrap/connection from frontend route tests.

## Migration Plan

1. Add BFF websocket proxy route and downstream rewrite to rt-gateway `/ws`.
2. Update ws-config payload generation to emit proxied wsUrl.
3. Update frontend realtime route/client tests and sequence docs.
4. Update runbook test flow evidence for proxied websocket ingress.
5. Rollback option: restore previous ws-config wsUrl direct target and disable proxy route mapping.

## Open Questions

- Should the gateway/BFF websocket proxy path be versioned under current `/v1/bff/mvp/realtime/*` family or promoted to a more generic `/v1/bff/realtime/*` route family in a future change?
