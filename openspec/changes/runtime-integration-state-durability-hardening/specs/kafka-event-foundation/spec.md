## ADDED Requirements

### Requirement: Runtime services MUST use real Kafka transport in runtime profiles
Services participating in jam queue, playback, and realtime fanout flows SHALL initialize real Kafka producers/consumers in runtime profiles and MUST NOT silently downgrade to noop transport.

#### Scenario: Runtime startup with valid Kafka connection
- **WHEN** `jam-service`, `playback-service`, or `rt-gateway` starts in runtime profile with valid broker configuration
- **THEN** each service initializes required Kafka clients successfully and marks transport as ready

#### Scenario: Runtime startup without valid Kafka configuration
- **WHEN** a service starts in runtime profile with missing or invalid Kafka broker configuration
- **THEN** startup fails fast with deterministic configuration error and service readiness remains false

### Requirement: Noop transport SHALL be restricted to explicit test profile
The system SHALL allow noop/mock Kafka transport only in explicit test profile and SHALL reject noop/mock transport selection in non-test runtime profiles.

#### Scenario: Noop transport requested in non-test profile
- **WHEN** a non-test runtime profile config selects noop/mock Kafka transport
- **THEN** service startup is rejected with deterministic invalid-runtime-adapter error
