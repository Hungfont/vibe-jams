---
description: "Use when running /opsx:apply, implementing OpenSpec tasks, or updating docs/runbooks/run.md. Enforces runbook initialization, continuous enrichment, and test-flow alignment with real execution behavior."
name: "Runbook Management Rules"
applyTo:
  - "docs/runbooks/run.md"
  - "openspec/changes/**/tasks.md"
  - ".github/commands/opsx-apply.md"
---
# Runbook Management Rules

- These rules are default preferences and should be followed unless the user explicitly requests a different approach.

## 1) Runbook Initialization

- If docs/runbooks/run.md is empty or missing structured flows, infer baseline flows from implemented code/tests/docs in the repository.
- Synthesize only reusable patterns that reflect actual execution behavior.
- Do not add speculative flows.

## 2) Continuous Enrichment

- Treat docs/runbooks/run.md as a living document.
- When introducing or changing execution flows, update existing sections before creating new ones.
- Consolidate duplicate flows into generalized patterns.

## 3) Integration with /opsx:apply

- For every new feature/task implemented via /opsx:apply, identify new or modified execution flows and update docs/runbooks/run.md.
- Maintain a dedicated Handbook / Test Flows section with:
  - Step-by-step flow
  - Expected outcome
  - Edge cases (when relevant)
- Ensure test flows match real behavior observed in implementation/tests.

## 4) Consistency and Quality

- Keep structure consistent with clear headings and ordered steps.
- Keep instructions concise, deterministic, and actionable.
- Prefer pattern-based guidance over one-off, task-specific notes.
- Avoid ambiguous language.

## 5) Agent Behavior Expectations

- Treat docs/runbooks/run.md as the single source of truth for execution flows.
- Before adding new logic, check for similar existing runbook flows and reuse/adapt first.
- If existing flows conflict, update runbook first, then implement.

## 6) Failure Handling and Assumptions

- If flows are unclear or inconsistent, infer the most likely pattern from current code/tests.
- Explicitly document assumptions in docs/runbooks/run.md so they can be refined later.
