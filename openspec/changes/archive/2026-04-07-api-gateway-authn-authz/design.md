## Context

Today the api-service calls `POST /internal/v1/auth/validate` on auth-service on every orchestration request, then forwards the raw `Authorization` header to jam-service and playback-service. There is no shared entry point — clients call api-service directly. The auth-service already exposes the internal validate endpoint (`POST /internal/v1/auth/validate`) which returns normalized claims (`userId`, `plan`, `sessionState`, `scope`). The goal is to move token validation to a gateway at the perimeter so downstream services operate with trusted identity context.

**Current flow:**
```
Client → api-service → auth-service /internal/v1/auth/validate
                     → jam-service (Authorization header forwarded)
                     → playback-service (Authorization header forwarded)
```

**Target flow:**
```
Client → api-gateway → auth-service /internal/v1/auth/validate (once)
       → api-gateway → api-service (X-Auth-* headers, no Authorization)
                     → jam-service (X-Auth-* headers forwarded)
                     → playback-service (X-Auth-* headers forwarded)
```

## Goals / Non-Goals

**Goals:**
- Single external ingress through api-gateway
- Token validation happens once at the gateway using the existing `/internal/v1/auth/validate` endpoint
- api-service and downstream services receive trusted `X-Auth-*` headers, never raw tokens
- api-service removes its direct auth-service call and reads identity from headers
- Public auth routes (`/v1/auth/**`) bypass token validation and route to auth-service

**Non-Goals:**
- Changes to auth-service API surface — `/internal/v1/auth/validate` is used as-is
- Fine-grained authZ for downstream services (jam-service, playback-service) — follow-up change
- Gateway rate limiting, circuit breaking, or load balancing — out of scope
- Changing the auth-service internal response contract — claims shape is unchanged
- mTLS between gateway and upstreams — header trust model is acceptable for now

## Decisions

### Decision 1: Use existing `/internal/v1/auth/validate` — no new auth-service endpoint

**Chosen**: api-gateway calls the existing `POST /internal/v1/auth/validate` exactly as api-service does today. The endpoint accepts `Authorization: Bearer <token>` and returns `{ userId, plan, sessionState, scope }`. No new endpoint needed.

**Rationale**: The endpoint already satisfies the gateway's needs — it validates the token, checks session revocation state, and returns normalized claims. Introducing a parallel endpoint would duplicate logic and split maintenance.

### Decision 2: `X-Auth-*` header convention for claim forwarding

**Chosen**: Gateway injects four headers after successful validation: `X-Auth-UserId`, `X-Auth-Plan`, `X-Auth-Scope`, `X-Auth-SessionState`. Values map directly from the validate response fields. The `Authorization` header is stripped. Any inbound client-supplied `X-Auth-*` headers are stripped before processing.

**Alternatives considered**:
- *Pass-through JWT*: Forward raw token — rejected; every downstream service would need a JWT library and the validate round-trip is already done at the gateway.
- *Single encoded header*: Base64 JSON blob — rejected; harder to consume and log; field-per-header is transparent and debuggable.

**Rationale**: Field-per-header is explicit, easy to read in logs, and easy to consume in Go with `r.Header.Get("X-Auth-UserId")`. The `X-Auth-` prefix is self-documenting and easy to strip on inbound.

### Decision 3: api-service removes direct auth-service call

**Chosen**: Remove `HTTPAuthClient` and the `authClient.ValidateBearerToken()` call from `service.Orchestrate()`. Replace with header extraction: `X-Auth-UserId`, `X-Auth-Plan`, `X-Auth-SessionState`, `X-Auth-Scope`. If `X-Auth-UserId` is missing or empty, return `401 missing_identity_context`. Validate `sessionState` remains `valid` using the same logic as today.

**Rationale**: Removes the redundant auth round-trip (gateway already did it). Simplifies api-service — no auth-service HTTP client needed. Forces the gateway to be the only auth path.

### Decision 4: api-service forwards `X-Auth-*` to downstream services

**Chosen**: Replace the current pattern (forwarding raw `Authorization` header to jam-service and playback-service) with forwarding the gateway-injected `X-Auth-*` headers. Downstream services receive identity context without raw tokens.

**Rationale**: Downstream services should never see raw client tokens. They only need identity fields to make authZ decisions. This removes the need for jam-service and playback-service to ever call auth-service themselves.

### Decision 5: Public route bypass list is static and gateway-owned

**Chosen**: The gateway bypasses token validation for `POST /v1/auth/login`, `POST /v1/auth/refresh`, `POST /v1/auth/logout`, `GET /v1/auth/me` — forwarding them directly to auth-service. All other routes go through authN middleware.

**Rationale**: Auth routes cannot require a valid token (bootstrapping problem). Static list is auditable. `GET /v1/auth/me` is included because it validates its own token internally in auth-service.

## Risks / Trade-offs

- **Header spoofing if api-service is reached directly**: If a client bypasses the gateway and reaches api-service directly, it could inject `X-Auth-*` headers. → Mitigation: Block external direct access to api-service at the network/firewall layer. Gateway is sole public ingress.
- **Gateway becomes critical path**: Every non-auth request flows through gateway → auth-service → api-service. → Mitigation: Deploy multiple gateway instances; auth-service already handles high validation load.
- **Added latency**: Gateway adds one HTTP hop. Previously api-service called auth-service directly — net latency change is zero for the auth call; overhead is only the gateway proxy itself.
- **Migration window**: During cutover, api-service retains a token-parsing fallback briefly. Must be removed after gateway is stable or it undermines the security model.

## Migration Plan

1. Merge auth-service with no changes (it's ready).
2. Deploy api-gateway, wired to auth-service and api-service. Shadow mode initially.
3. Add `X-Auth-*` header extraction to api-service alongside existing auth-service call (feature-flag the fallback).
4. Cut public traffic over: update ingress/DNS to route through gateway.
5. Remove direct auth-service call and fallback from api-service once gateway is confirmed healthy.
6. Block external direct access to api-service at the network layer.

## Open Questions

- Should api-gateway cache validate responses (keyed on token hash) to reduce auth-service load? If yes, what TTL is acceptable given logout revocation latency?
- Do jam-service and playback-service need to validate the `X-Auth-*` headers they receive, or is implicit trust from api-service sufficient for this phase?
