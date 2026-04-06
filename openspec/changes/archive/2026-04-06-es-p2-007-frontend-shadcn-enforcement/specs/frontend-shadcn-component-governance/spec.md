## ADDED Requirements

### Requirement: Frontend primitive layer SHALL enforce shadcn/ui-first component usage
The frontend system SHALL treat shadcn/ui primitives as the default primitive layer and MUST reject introducing duplicate custom primitives when an equivalent approved primitive exists.

#### Scenario: Equivalent primitive already approved
- **WHEN** a contributor introduces a new primitive with behavior equivalent to an approved shadcn primitive category
- **THEN** the validation flow fails and requires reuse of the approved primitive implementation

#### Scenario: No equivalent primitive exists
- **WHEN** a required primitive has no approved shadcn equivalent in the project
- **THEN** implementation may proceed only with an explicit documented exception and rationale

### Requirement: Frontend validation flow SHALL include deterministic primitive conformance checks
The frontend lint/CI pipeline SHALL execute primitive conformance checks and MUST fail deterministically for policy violations.

#### Scenario: Violation introduced in pull request
- **WHEN** new code introduces a non-compliant primitive pattern
- **THEN** the frontend validation step returns a failing status with actionable error detail

#### Scenario: Conformant change
- **WHEN** code composes UI from approved primitive inventory
- **THEN** conformance checks pass without additional exception requirements

### Requirement: Primitive exceptions SHALL be governed by explicit approval metadata
The system SHALL require each exception to include owner, rationale, and review status so deviations are auditable.

#### Scenario: Exception missing metadata
- **WHEN** an exception entry omits required approval metadata
- **THEN** the validation flow fails and requests complete exception details

#### Scenario: Approved exception entry present
- **WHEN** an exception includes required metadata and approved status
- **THEN** the validation flow allows the scoped deviation only for that declared primitive case
