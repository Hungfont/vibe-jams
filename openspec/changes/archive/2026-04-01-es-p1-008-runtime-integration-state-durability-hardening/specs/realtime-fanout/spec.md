## ADDED Requirements

### Requirement: Websocket handshake MUST enforce configured origin allowlist
rt-gateway SHALL validate websocket handshake Origin against configured allowlist and SHALL reject unknown origins by default.

#### Scenario: Allowlisted origin succeeds
- **WHEN** a websocket client connects with Origin that exists in configured allowlist
- **THEN** handshake succeeds and connection is accepted

#### Scenario: Unknown origin is rejected
- **WHEN** a websocket client connects with Origin that is not allowlisted
- **THEN** handshake is rejected with deterministic forbidden-origin behavior and connection is closed
