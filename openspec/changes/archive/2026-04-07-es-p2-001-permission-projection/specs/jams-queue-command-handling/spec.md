## ADDED Requirements

### Requirement: Queue reorder authorization SHALL enforce projected guest permissions
The `jams` service command handling path SHALL allow reorder commands from host actors and from guest actors only when `canReorderQueue=true` in session permission projection state.

#### Scenario: Guest reorder denied when capability disabled
- **WHEN** guest actor submits reorder command while `canReorderQueue` is disabled
- **THEN** command is rejected with deterministic forbidden permission error and queue state remains unchanged

#### Scenario: Guest reorder allowed when capability enabled
- **WHEN** guest actor submits reorder command while `canReorderQueue` is enabled and optimistic version checks pass
- **THEN** reorder mutation is applied with normal version increment semantics