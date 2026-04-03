# Copilot Instructions

## Agent Selection

Use the correct specialist agent based on task domain.

### Frontend Agent
- Agent: Frontend Engineer
- Agent file: .github/agents/frontend.agent.md
- Use for: UI components, pages, styling, responsiveness, frontend integration and tests.
- Code scope: only update frontend/** unless user explicitly asks for override.
- Rules scope: read only .github/rules/frontend/** when that folder has relevant files.
- Instruction scope: use only .github/instructions/frontend/fe-*.md unless user explicitly asks for override.

### Backend Agent
- Agent: Backend Engineer
- Agent file: .github/agents/backend.agent.md
- Use for: backend APIs, handlers, repositories, Kafka, config, and backend tests.
- Code scope: only update backend/** unless user explicitly asks for override.
- Rules scope: read only .github/rules/backend/**.
- Instruction scope: use only .github/instructions/be-*.md unless user explicitly asks for override.

## Code Update Policy

- Only update code relevant to the selected agent.
- Do not mix frontend and backend edits in one pass unless the user explicitly requests cross-domain changes.
- If the request is ambiguous, choose the smallest relevant scope first and keep changes minimal.
- If cross-domain work is required, state boundary and update each domain deliberately.

## Safety and Quality

- Keep existing public contracts stable unless change is requested.
- Prefer small, focused patches over broad refactors.
- Run relevant validation for the edited domain before completion.
