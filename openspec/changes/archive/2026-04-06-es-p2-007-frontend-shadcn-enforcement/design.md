## Context

Frontend already uses a `components/ui` layer, but primitive implementations are not consistently aligned with shadcn conventions and there is no automated enforcement to prevent drift. Current policy is documented in instruction files, yet lint/CI does not fail when duplicate custom primitives are introduced. This change is frontend-scoped and must preserve existing page behavior while tightening UI primitive governance.

## Goals / Non-Goals

**Goals:**
- Enforce shadcn/ui-first primitive usage for frontend components when an equivalent exists.
- Add deterministic validation so CI rejects newly introduced duplicate primitive patterns.
- Define an approved primitive inventory and exception mechanism for cases without shadcn equivalents.
- Provide a migration path that upgrades existing primitives incrementally without blocking unrelated delivery.

**Non-Goals:**
- Rebuild every existing UI component in one pass.
- Introduce a new UI framework or replace Tailwind.
- Change backend contracts, routes, or runtime behavior.

## Decisions

1. Policy scope alignment to frontend source files
- Decision: enforce policy over `frontend/src/**/*.{ts,tsx,js,jsx}` so it covers actual component files.
- Rationale: most UI code is in `.tsx`; narrower scope leaves enforcement gaps.
- Alternative considered: keep broad repository-level glob. Rejected because it adds noise outside frontend domain.

2. Add conformance checks to frontend validation flow
- Decision: add deterministic checks in frontend lint pipeline to detect non-approved primitive additions.
- Rationale: policy-only guidance is insufficient; CI must be the source of enforcement truth.
- Alternative considered: manual code review only. Rejected due to inconsistency and reviewer load.

3. Primitive inventory + exception registry
- Decision: define approved primitive set (`button`, `input`, `card`, `dialog`, etc.) and require explicit documented exception for custom additions.
- Rationale: allows controlled evolution while preserving governance.
- Alternative considered: blanket prohibition of all custom files under `components/ui`. Rejected because some missing primitives may still require local implementation.

4. Incremental migration strategy
- Decision: enforce no-new-violations first, then migrate existing duplicate primitives in prioritized batches.
- Rationale: avoids large risky refactor while still moving toward full conformance.
- Alternative considered: big-bang migration. Rejected due to regression risk and delivery disruption.

## Risks / Trade-offs

- [Risk: false positives block merges] -> Mitigation: maintain explicit allowlist/exception file with owner and rationale.
- [Risk: migration churn across many components] -> Mitigation: phased migration with compatibility wrappers where needed.
- [Risk: perceived slowdown from stricter lint gates] -> Mitigation: limit initial checks to high-signal primitive categories and improve developer guidance in errors.
- [Risk: policy drift between instruction docs and lint implementation] -> Mitigation: keep one canonical inventory and reference it from both instruction and lint checks.

## Migration Plan

1. Finalize approved primitive inventory and exception workflow.
2. Enable frontend conformance checks in lint/CI with non-blocking dry run to baseline violations.
3. Switch checks to blocking for new violations.
4. Migrate existing duplicate primitives in batches and retire deprecated implementations.
5. Update runbook validation flow with the new shadcn conformance step.

Rollback strategy:
- Temporarily downgrade conformance check from blocking to warning.
- Keep migration commits isolated so specific primitive migrations can be reverted without affecting page flows.

## Open Questions

- Should exception approvals be tracked in a dedicated YAML/JSON registry or in markdown runbook entries?
- Which primitive categories are mandatory in phase 1 enforcement (`button/input/card/dialog/toast`) versus later phases?
- Should conformance checks fail on file creation only, or also on forbidden import usage in feature components?
