## ADDED Requirements

### Requirement: Frontend websocket ingress SHALL be mediated by gateway/BFF proxy path
The rt-gateway websocket `/ws` endpoint SHALL be treated as downstream infrastructure ingress for BFF-proxied frontend connections, and frontend contracts SHALL use gateway/BFF websocket path for connection establishment.

#### Scenario: Proxied websocket ingress reaches rt-gateway
- **WHEN** frontend opens websocket via gateway/BFF route with valid session query
- **THEN** rt-gateway accepts the proxied websocket handshake and applies the same room subscription/fanout semantics

#### Scenario: Direct frontend websocket target is non-compliant
- **WHEN** frontend implementation targets rt-gateway `/ws` directly instead of gateway/BFF websocket path
- **THEN** implementation is non-compliant with realtime ingress policy
