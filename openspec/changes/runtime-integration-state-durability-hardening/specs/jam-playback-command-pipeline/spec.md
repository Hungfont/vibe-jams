## ADDED Requirements

### Requirement: Playback state SHALL be durable and restart-recoverable
The playback command pipeline SHALL persist accepted playback transitions in durable storage so effective playback state is recoverable after restart.

#### Scenario: Accepted command state recovered after restart
- **WHEN** a host playback command is accepted and the service restarts
- **THEN** the recovered playback state reflects the last committed transition and command sequencing continues from persisted version

### Requirement: Durability commit MUST precede transition publish
The playback pipeline SHALL persist transition state before publishing outbound playback transition events.

#### Scenario: Publish follows successful durable commit
- **WHEN** a playback transition command is accepted in runtime profile
- **THEN** the service commits transition state durably before publishing the corresponding playback event
