## Why

The platform has no centralized authentication perimeter. Every service that needs identity (api-service, jam-service, etc.) calls `auth-service /internal/v1/auth/validate` directly on each request — duplicating the auth round-trip, coupling services to auth-service, and making it impossible to enforce a consistent auth policy. An api-gateway layer is needed to validate tokens once at the edge and forward normalized identity claims downstream, so internal services operate with trusted context instead of raw tokens.

## What Changes

- **New**: `api-gateway` service — single external ingress for all client traffic
- **New**: api-gateway validates Bearer tokens by calling the existing `POST /internal/v1/auth/validate` on auth-service, then injects `X-Auth-UserId`, `X-Auth-Plan`, `X-Auth-Scope`, `X-Auth-SessionState` headers into forwarded requests
- **New**: api-gateway strips the `Authorization` header and client-supplied `X-Auth-*` headers before forwarding to upstream services
- **Modified**: `api-service` removes its direct auth-service call; reads identity from gateway-forwarded `X-Auth-*` headers instead
- **Modified**: `api-service` forwards `X-Auth-*` headers to downstream services (jam-service, playback-service) instead of the raw `Authorization` header
- **No change**: auth-service API surface — `POST /internal/v1/auth/validate` already exists and is sufficient; no new endpoints needed

## Capabilities

### New Capabilities

- `api-gateway-routing`: Gateway receives all client requests; routes auth paths (`/v1/auth/**`) directly to auth-service (public bypass); routes all other paths to api-service after applying authN middleware; enforces per-request upstream timeout
- `api-gateway-authn-middleware`: Gateway validates Bearer tokens using the existing `POST /internal/v1/auth/validate`; strips inbound `X-Auth-*` headers to prevent spoofing; on success injects `X-Auth-UserId`, `X-Auth-Plan`, `X-Auth-Scope`, `X-Auth-SessionState` and strips `Authorization` before forwarding

### Modified Capabilities

- `api-service-bff-orchestration`: BFF orchestration MUST read identity from gateway-forwarded `X-Auth-*` headers instead of calling auth-service directly; MUST forward `X-Auth-*` headers to downstream services (jam-service) instead of the raw `Authorization` header
- `auth-claim-contract`: The normalized claim contract (`userId`, `plan`, `scope`, `sessionState`) now applies to both the existing auth validation flows and to gateway-injected `X-Auth-*` headers consumed by api-service; all surfaces MUST use the same field names and semantics

## Impact

- **api-gateway** (new Go service): Sits in front of api-service; requires outbound connectivity to auth-service; must be the only public ingress — direct external access to api-service is blocked at the network layer
- **api-service**: Removes `HTTPAuthClient` dependency and direct auth-service call; replaces with header extraction from `X-Auth-*`; forwards same headers to jam-service and playback-service
- **auth-service**: No API changes; `POST /internal/v1/auth/validate` continues to serve both gateway (new caller) and any remaining direct service callers during migration
- **jam-service / playback-service**: Receive `X-Auth-*` headers instead of `Authorization`; no validation changes required in this change — they will be addressed in a follow-up authZ change
- **Deployment**: Gateway must be the only public ingress; api-service direct external access must be blocked at the firewall/ingress level
