## Why

Current frontend server routes still call multiple microservices directly for jam, playback, catalog, and realtime bootstrap. This creates inconsistent service-entry behavior and bypasses centralized BFF policy controls. We need one mandatory path where non-auth microservice HTTP calls always go through api-service (BFF) first.

## What Changes

- Enforce a mandatory BFF-first routing rule for non-auth microservice HTTP flows: frontend `/api/**` routes -> api-gateway -> api-service (BFF) -> downstream microservices.
- Keep auth lifecycle and explicit authN/authZ endpoints as dedicated auth flows.
- Extend BFF entrypoint surface to handle command/query routing currently sent directly to jam-service, playback-service, and catalog-service.
- Route realtime bootstrap config (`/api/realtime/ws-config`) through BFF before returning websocket connection settings.
- Update sequence documentation and route mapping to reflect the mandatory BFF hop model.

## Capabilities

### New Capabilities
- `api-service-bff-microservice-routing`: defines BFF endpoints, identity propagation, downstream delegation, and deterministic envelope mapping for jam/playback/catalog/realtime bootstrap calls.

### Modified Capabilities
- `frontend-phase1-ui-routing-and-flows`: updates route-to-service flow so non-auth microservice HTTP calls and realtime bootstrap are BFF-first.
- `api-service-bff-orchestration`: expands BFF role from orchestration-only entrypoint to broader frontend microservice routing gateway for non-auth HTTP operations.

## Impact

- Frontend API routes under `frontend/src/app/api/**` for jam, playback, catalog, and realtime bootstrap.
- api-service BFF handlers/routers and downstream clients.
- api-gateway route forwarding expectations for BFF endpoints.
- Documentation: `docs/frontend-backend-sequence.md`, `docs/runbooks/run.md`.
- Test suites for frontend route forwarding, BFF routing, and integration flow parity.
