## Why

Frontend currently mixes handcrafted primitive implementations with shadcn-style usage, so the "shadcn-first" policy is guidance but not enforceable behavior. This creates UI drift, inconsistent accessibility defaults, and repeated styling patterns that increase maintenance cost.

## What Changes

- Introduce enforceable frontend governance that requires shadcn/ui primitives for equivalent component categories.
- Add deterministic validation in frontend CI/lint flow to detect and reject newly introduced duplicate primitive components.
- Define an approved primitive inventory and migration policy for existing custom primitives to either shadcn-generated components or documented exceptions.
- Standardize import and composition patterns for feature components so they compose from approved primitives only.
- Add explicit exception workflow (documented approval + rationale) when no shadcn equivalent exists.

## Capabilities

### New Capabilities
- `frontend-shadcn-component-governance`: Enforce shadcn/ui-first primitive usage with policy checks, exception handling, and migration guardrails.

### Modified Capabilities
- `frontend-phase1-ui-routing-and-flows`: Frontend UI flow requirements are tightened so page/component implementations must consume approved primitive layer contracts.

## Impact

- Affected code:
  - `frontend/src/components/ui/**`
  - `frontend/src/components/**`
  - `frontend/eslint.config.mjs`
  - `frontend/package.json` scripts and validation workflow
- Affected docs/policy:
  - `.github/instructions/frontend/fe-shadcn-ui-only.instructions.md`
  - `docs/runbooks/run.md` (frontend validation flow)
- Affected CI behavior:
  - Frontend lint/test/build gating includes shadcn primitive conformance checks.
