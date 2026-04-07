## 1. Websocket Proxy Implementation

- [x] 1.1 Add api-service BFF websocket proxy route for realtime connect path and rewrite to rt-gateway `/ws` with query passthrough.
- [x] 1.2 Update api-service router registration to expose the websocket proxy route alongside ws-config route.
- [x] 1.3 Update realtime bootstrap ws-config payload to return gateway/BFF websocket URL instead of direct rt-gateway URL.

## 2. Contract and Documentation Alignment

- [x] 2.1 Update api-service and api-gateway OpenAPI docs to include websocket proxy path and bootstrap contract behavior.
- [x] 2.2 Update `docs/frontend-backend-sequence.md` realtime sequence and route mapping to reflect websocket proxy ingress.
- [x] 2.3 Update `docs/runbooks/run.md` with websocket-through-BFF execution and validation flow.

## 3. Frontend and Test Validation

- [x] 3.1 Update frontend realtime ws-config route/tests so returned `wsUrl` targets gateway/BFF websocket path.
- [x] 3.2 Add/update backend tests for api-service websocket proxy routing and ws-config contract.
- [x] 3.3 Run focused backend tests (`api-service`, `api-gateway`) and frontend realtime route tests.
- [x] 3.4 Run strict OpenSpec validation for `bff-websocket-proxy-routing`.
