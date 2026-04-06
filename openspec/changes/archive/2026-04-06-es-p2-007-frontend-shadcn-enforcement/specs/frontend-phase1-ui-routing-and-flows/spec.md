## MODIFIED Requirements

### Requirement: Phase 1 UI SHALL use shadcn components to provide Spotify-like interaction structure
The frontend SHALL compose lobby and room surfaces using approved shadcn primitives and Tailwind utilities to achieve Spotify-like information hierarchy and interaction behavior, and MUST enforce this policy through frontend validation checks for new code.

#### Scenario: Lobby composition
- **WHEN** rendering lobby create/join page
- **THEN** UI uses approved shadcn Card, Tabs, Input, Button, Alert, and Toast primitives for action and feedback

#### Scenario: Jam room composition
- **WHEN** rendering jam room and focused views
- **THEN** UI uses approved shadcn Card, Tabs, Badge, ScrollArea, Slider, Tooltip, Dialog, Skeleton, and Alert patterns for queue, playback, participants, and diagnostics

#### Scenario: New duplicate primitive introduced
- **WHEN** a change introduces a custom primitive duplicating an approved shadcn-equivalent primitive category
- **THEN** frontend conformance validation fails and requires primitive reuse or approved exception
