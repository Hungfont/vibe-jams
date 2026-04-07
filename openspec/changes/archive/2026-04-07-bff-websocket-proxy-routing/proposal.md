## Why

Realtime bootstrap already goes through BFF, but websocket traffic still connects directly to rt-gateway `/ws`, bypassing the intended BFF-first ingress policy for frontend-to-microservice communication. We need websocket connection settings and websocket access path to route through api-gateway -> api-service (BFF) so frontend no longer calls rt-gateway directly.

## What Changes

- Add a BFF websocket proxy route family for realtime connections so browser websocket traffic enters through api-gateway and api-service before reaching rt-gateway.
- Update realtime bootstrap response contract so returned `wsUrl` targets the gateway/BFF websocket path instead of direct rt-gateway `/ws`.
- Keep existing realtime bootstrap HTTP flow (`/api/realtime/ws-config`) through BFF and align it with the new proxied websocket connect path.
- Update frontend realtime expectations and sequence docs to disallow direct browser websocket calls to rt-gateway.
- Add tests for proxy route behavior, wsUrl contract, and frontend route mapping parity.

## Capabilities

### New Capabilities
- `api-service-bff-websocket-proxy`: defines BFF websocket proxy behavior for realtime connect path and ws bootstrap contract mapping.

### Modified Capabilities
- `frontend-phase1-ui-routing-and-flows`: realtime path changes from direct rt-gateway websocket connect to gateway/BFF-mediated websocket connect.
- `realtime-fanout`: clarifies rt-gateway websocket endpoint usage as internal downstream of BFF/gateway proxy path for frontend clients.

## Impact

- Backend routing/proxy logic in api-service BFF for websocket upgrade proxying to rt-gateway.
- Frontend realtime bootstrap contract handling under `frontend/src/app/api/realtime/ws-config` and jam room realtime client connect behavior.
- OpenAPI/Swagger docs for realtime websocket proxy route families.
- Documentation updates in `docs/frontend-backend-sequence.md` and `docs/runbooks/run.md`.
- Focused backend/frontend tests for ws proxy path and bootstrap contract correctness.
