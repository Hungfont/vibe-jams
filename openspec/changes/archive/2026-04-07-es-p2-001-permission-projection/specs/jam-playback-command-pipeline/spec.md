## MODIFIED Requirements

### Requirement: Authorization and host-only enforcement
The system SHALL reject playback commands when authentication is missing/invalid, when session state is not active, or when actor lacks playback control permission for the target session. Host actors are always authorized. Guest actors SHALL require `canControlPlayback=true` in projected permission state.

#### Scenario: Missing or invalid authentication
- **WHEN** a playback command request is sent without valid authentication context
- **THEN** the system returns an unauthorized error response

#### Scenario: Authenticated guest command attempt without permission
- **WHEN** an authenticated guest without playback permission sends a playback command for a session
- **THEN** the system returns a deterministic forbidden permission error response

#### Scenario: Authenticated guest command attempt with permission
- **WHEN** an authenticated guest with playback permission sends a playback command for a session
- **THEN** the system authorizes command execution and applies normal validation/processing semantics