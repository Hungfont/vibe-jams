## ADDED Requirements

### Requirement: Websocket handshake SHALL enforce origin allowlist
`rt-gateway` SHALL validate websocket handshake `Origin` against configured allowlist and SHALL reject unknown origins by default.

#### Scenario: Allowlisted origin connects successfully
- **WHEN** a websocket client connects with an `Origin` value present in configured allowlist
- **THEN** handshake succeeds and connection proceeds to subscription flow

#### Scenario: Unknown origin is rejected
- **WHEN** a websocket client connects with an `Origin` value not present in configured allowlist
- **THEN** handshake is rejected with deterministic forbidden-origin response and connection is closed

### Requirement: Fanout runtime path SHALL require real Kafka consumer integration
`rt-gateway` fanout path SHALL initialize real Kafka consumer integration in runtime profiles and MUST NOT operate in noop fanout mode outside explicit test profile.

#### Scenario: Fanout startup rejects noop mode in runtime profile
- **WHEN** runtime profile selects noop/mock fanout consumer
- **THEN** startup fails with deterministic invalid-runtime-adapter error and fanout is not started
