---
description: Docs workflow for US to ES expansion and engineer-story file split
alwaysApply: false
---

# Docs Story Workflow

## Scope
- This rule is manual-only.
- Apply this workflow only when explicitly invoked with `@docs-story-workflow`.
- Do not auto-apply this rule based on file paths or inferred scope.

## Story authoring rules
- Keep existing section format and numbering style in `docs/story/user-story.md`.
- Add only minimum unblock stories for `auth-service` and `catalog-service` per phase.
- Keep acceptance criteria concise, testable, and aligned with API/error behavior.

## US to ES mapping rules
- For each new user story, add mapped engineer stories in `docs/story/enginner-story.md`.
- Keep ES sections in the same format: title, task description, dependencies, definition of done.
- Ensure ES content is implementation-focused and traceable back to the corresponding US.

## Per-ES file rules
- Create one file per ES in `docs/story/engineer-stories/`.
- Use naming format `[ES-ID].[Phase ...].md` exactly.
- Do not rename existing ES files.
- Keep file content consistent with the ES section in `enginner-story.md`.

## Validation checklist
- New US and ES IDs are unique and phase-correct.
- Every newly added ES has a corresponding file in `docs/story/engineer-stories/`.
- Naming and markdown format remain consistent with existing files.
