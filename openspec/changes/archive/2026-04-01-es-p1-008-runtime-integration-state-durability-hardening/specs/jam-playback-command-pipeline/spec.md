## ADDED Requirements

### Requirement: Playback runtime state SHALL be durably persisted
Accepted playback transitions SHALL be committed to durable runtime storage before command completion in non-test profiles.

#### Scenario: Playback transition persists across restart
- **WHEN** a playback transition command is accepted and service restarts
- **THEN** recovered playback state reflects the latest committed transition and sequence metadata

#### Scenario: Durable commit precedes event publication
- **WHEN** a playback transition is accepted in runtime mode
- **THEN** persistent state commit occurs before publishing transition event to Kafka
