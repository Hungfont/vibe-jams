# Copilot Instructions

## Agent Selection

Use the correct specialist agent based on task domain.
- BEFORE implementation, proposal, explore, apply, read only relevant skills first. After that load rules, instructions in .github.
- DO NOT load or apply skills files outside .github/skills/**, instructions files outside .github/instructions/**, rules files outside .github/rules/** unless explicitly requested.

### Frontend Agent
- Agent: Frontend Engineer
- Agent file: .github/agents/frontend.agent.md
- Use for: UI components, pages, styling, responsiveness, frontend integration and tests.
- Code scope: only update frontend/** unless user explicitly asks for override.
- Rules scope: read only .github/rules/frontend/** when that folder has relevant files.
- Instructions scope: use instructions from .github/instructions/frontend/fe-*.instructions.md unless user explicitly asks for override.
- Skills scope: use skills from .github/skills/** when that folder has relevant files and only if relevant to the task at hand.

### Backend Agent
- Agent: Backend Engineer
- Agent file: .github/agents/backend.agent.md
- Use for: backend APIs, handlers, repositories, Kafka, config, and backend tests.
- Code scope: only update backend/** unless user explicitly asks for override.
- Rules scope: read only .github/rules/backend/**.
- Instruction scope: use only .github/instructions/backend/be-*.instructions.md unless user explicitly asks for override.
- Skills scope: use skills from .github/skills/** when that folder has relevant files and only if relevant to the task at hand.

## Code Update Policy

- Only update code relevant to the selected agent.
- Do not mix frontend and backend edits in one pass unless the user explicitly requests cross-domain changes.
- If the request is ambiguous, choose the smallest relevant scope first and keep changes minimal.
- If cross-domain work is required, state boundary and update each domain deliberately.

## Safety and Quality

- Keep existing public contracts stable unless change is requested.
- Prefer small, focused patches over broad refactors.
- Run relevant validation for the edited domain before completion.

## OpenSpec Setup By Custom Agent

When running `/opsx:explore`, `/opsx:propose`, `/opsx:apply`, or `/opsx:archive`, select one execution agent first and load resources from the mapped paths below.

### 1) Frontend Engineer
- Trigger: change scope is primarily `frontend/**`, UI flows, Next.js routing, styling, or frontend integration tests.
- Rules: `.github/rules/frontend/**` (optionally `.github/rules/common/**` for shared security or patterns).
- Instructions: `.github/instructions/frontend/fe-*.instructions.md`.
- Skills: `.github/skills/nextjs-app-router-patterns/SKILL.md`, `.github/skills/tailwind-design-system/SKILL.md`, `.github/skills/tdd-workflow/SKILL.md` when relevant.
- Prompts: `.github/prompt/frontend/*.prompt.md` when generating or updating frontend instruction assets.

### 2) Backend Engineer
- Trigger: change scope is primarily `backend/**`, APIs, handlers, repositories, Kafka, config, or backend tests.
- Rules: `.github/rules/backend/**` (optionally `.github/rules/common/**` for shared security or patterns).
- Instructions: `.github/instructions/backend/be-*.instructions.md`.
- Skills: `.github/skills/golang-backend-development/SKILL.md`, `.github/skills/kafka-engineer/SKILL.md`, `.github/skills/redis-best-practices/SKILL.md`, `.github/skills/tdd-workflow/SKILL.md` when relevant.
- Prompts: use workspace prompts if backend prompt assets are added later.

### 3) Agent (default / orchestration)
- Trigger: OpenSpec artifact work (`openspec/**`), cross-domain planning, archive or workflow coordination.
- Rules: `.github/rules/common/**` and domain rules only for files being changed.
- Instructions: load only domain instructions tied to impacted files.
- Skills: `.github/skills/openspec-explore/SKILL.md`, `.github/skills/openspec-propose/SKILL.md`, `.github/skills/openspec-apply-change/SKILL.md`, `.github/skills/openspec-archive-change/SKILL.md`.
- Prompts: use domain prompts only when generating prompt-driven artifacts.

### Execution policy
- Do not mix frontend and backend implementation in one pass unless explicitly requested.
- If OpenSpec tasks span both domains, split implementation by agent/domain and report boundaries clearly.
