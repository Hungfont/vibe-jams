## ADDED Requirements

### Requirement: Runtime transport MUST use real Kafka in non-test profiles
Services participating in jam queue, playback, and realtime fanout SHALL initialize real Kafka clients in non-test profiles and MUST reject noop transport configuration.

#### Scenario: Non-test runtime with real broker configuration
- **WHEN** jam-service, playback-service, or rt-gateway starts in a non-test profile with valid broker settings
- **THEN** required Kafka producer and consumer clients are initialized and runtime is marked ready

#### Scenario: Non-test runtime with noop transport configured
- **WHEN** a non-test profile attempts to start with noop or mock Kafka transport
- **THEN** startup fails with deterministic invalid-runtime-adapter error and service readiness stays false

### Requirement: Local profile SHALL provide deterministic broker fallback
The system SHALL apply deterministic localhost broker defaults in local profile when explicit broker endpoints are not provided.

#### Scenario: Local profile without explicit broker endpoint
- **WHEN** a local runtime starts without service-specific broker endpoint configuration
- **THEN** startup uses documented localhost broker fallback and logs effective configuration source

### Requirement: Shared noop transport SHALL be reused for non-default Kafka services
Services that do not use Kafka as default transport SHALL reuse shared `NoOpsProducer` and `NoOpsConsumer` from the shared module instead of service-local noop implementations.

#### Scenario: Service boot with non-Kafka default transport
- **WHEN** a service starts in a profile where Kafka transport is not configured as default
- **THEN** producer and consumer fallback paths use shared noop transport implementations

#### Scenario: Consistent noop behavior across services
- **WHEN** multiple services run with noop transport fallback
- **THEN** behavior and interfaces are consistent because they use shared noop transport components
