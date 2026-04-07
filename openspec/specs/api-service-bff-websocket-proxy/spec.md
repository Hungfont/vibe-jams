# api-service-bff-websocket-proxy Specification

## Purpose
TBD - created by archiving change bff-websocket-proxy-routing. Update Purpose after archive.
## Requirements
### Requirement: API-service BFF MUST proxy frontend websocket connect path to rt-gateway
API-service BFF SHALL expose a websocket proxy route family for frontend realtime connect requests and SHALL forward websocket upgrade traffic to rt-gateway `/ws` while preserving query parameters.

#### Scenario: Websocket connect path proxied through BFF
- **WHEN** a frontend client connects to the gateway-facing BFF websocket route with `sessionId` and optional `lastSeenVersion`
- **THEN** api-service proxies the request to rt-gateway `/ws` and preserves query values required by realtime fanout session recovery

#### Scenario: Unsupported method on websocket proxy route
- **WHEN** a non-websocket-compatible method is sent to the websocket proxy route
- **THEN** api-service returns deterministic method-not-allowed behavior

### Requirement: Realtime bootstrap config MUST return proxied websocket URL
The BFF realtime bootstrap response SHALL return `wsUrl` that targets the gateway/BFF websocket proxy route rather than direct rt-gateway `/ws`.

#### Scenario: Bootstrap response returns gateway/BFF websocket target
- **WHEN** frontend requests realtime bootstrap config through BFF
- **THEN** response includes `wsUrl` pointing to gateway/BFF websocket route plus normalized `sessionId` and `lastSeenVersion`

