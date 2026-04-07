---
name: /fullstack
id: fullstack
category: Workflow
description: Orchestrate OpenSpec-driven fullstack delivery — resolve change, run propose, gate approval, then apply.
---

Run a fullstack OpenSpec-driven workflow from requirement review to integrated implementation.

**Input**: Optionally provide a change name or a short feature request after `/fullstack`.

**IMPORTANT**: Do NOT guess or auto-select a change. Always let the user choose.

<!-- ALWAYS project flows will apply the strictly flow with #Sequence1: HTTP Request/Response Flow in `docs/frontend-backend-sequence.md` files and never update this sequence for any request unless explicitly requested. -->

**Primary path (always)**: `/opsx-propose` → approval gate → `/opsx:apply` (then `/opsx:archive` when complete).

## Workflow

1. **Resolve target change**

   - If user provided a change name, use it.
   - Otherwise run:
     ```bash
     openspec list --json
     ```
   - If exactly one relevant active change exists, use it.
   - If multiple are possible, use the **AskUserQuestion tool** to ask the user to select.

2. **Review OpenSpec coverage for the requested feature**

   - Check current status:
     ```bash
     openspec status --change "<name>" --json
     ```
   - Read relevant existing artifacts (proposal, design, specs, tasks) if present.
   - Classify outcome:
     - **A. Additive requirement only**: existing feature exists, only extra acceptance/constraint is needed.
     - **B. Existing feature gap**: feature exists but required behavior is missing in artifacts.
     - **C. Missing feature**: no relevant requirement/feature exists in OpenSpec.

3. **Run `/opsx-propose` (always required before apply)**

   For **A**, **B**, or **C** — and for any completed change that needs updates:
   - Run `/opsx-propose <name>` (or `/opsx-propose <new-change-name>` for case C or completed changes).
   - Do not replicate artifact creation steps here — `/opsx-propose` owns that workflow entirely.
   - Wait for `/opsx-propose` to complete before proceeding to the approval gate.

4. **User concern and approval gate (hard stop — mandatory before apply)**

   Immediately after `/opsx-propose` completes, present a concise summary and request explicit user decision:
   - OpenSpec artifacts updated or created
   - Expected backend changes
   - Expected frontend/integration changes
   - Potential contract or flow impact

   Use the **AskUserQuestion tool** to collect an explicit decision:
   - Approve and continue to `/opsx:apply <name>`
   - I have concerns (stay in proposal phase)

   Treat any response that is not explicit approval as **not approved**.

   If the user has concerns, stay in proposal phase and offer refinement options:
   1. Tighten or rewrite requirements in proposal
   2. Adjust acceptance criteria in specs
   3. Update design constraints and boundaries
   4. Split scope into a follow-up change
   5. Keep current scope but document risks/assumptions
   6. (Allow free-form user input)

   Apply chosen refinements to proposal/design/spec/tasks, then re-run the concern check until explicit approval is received. Do not run `/opsx:apply` without explicit approval.

5. **Run `/opsx:apply <name>`**

   - Proceed only after explicit user approval.
   - Run `/opsx:apply <name>` — do not replicate agent setup, task loop, or backend/frontend implementation steps here. `/opsx:apply` owns that workflow entirely.
   - Wait for `/opsx:apply` to complete or pause before proceeding.

6. **Integration and flow verification**

   After apply completes, validate end-to-end behavior for changed flows:
   - Frontend request/response types and adapters match backend contract.
   - Frontend no longer sends removed/deprecated fields.
   - Frontend handling for backend success/error envelope stays consistent.

   If flow changed intentionally, update documentation in the same change:
   - `docs/frontend-backend-sequence.md`
   - `docs/runbooks/run.md`

7. **Suggest next action**

   - If all tasks are done: suggest `/opsx:archive`.
   - If tasks remain: suggest `/opsx:apply <name>` to continue.
   - If new requirements surface: suggest `/opsx-propose <name>` to refresh artifacts.

## Guardrails

- Never skip OpenSpec update when requirement changes are discovered.
- Primary path is always: `/opsx-propose` → approval gate → `/opsx:apply`.
- Do not re-implement artifact creation, agent context setup, or task execution — delegate fully to `/opsx-propose` and `/opsx:apply`.
- Mandatory approval gate: explicit user approval required before `/opsx:apply`, immediately after propose output is presented.
- Hard stop rule: if approval is pending or concerns remain, continue proposal refinement — do not enter apply.
- If user raises concerns, remain in proposal phase and refine artifacts before any implementation.
- If requirements are unclear, ask for clarification before coding.
- Ensure runbook and sequence docs reflect real behavior when changed.
- Verify frontend-backend integration alignment before completion.

## Output Template

```md
## Fullstack Execution Summary

**Change:** <name>
**Coverage Result:** A | B | C
**Primary Path:** /opsx-propose → approval gate → /opsx:apply (then /opsx:archive when complete)
**Approval Status:** Approved | Pending concerns
**Approval Checkpoint:** Requested after propose | Not requested (invalid)

### Concerns and Resolution
- User concerns: <none or summary>
- Selected option(s): <1–6 or custom>
- Proposal-phase refinements: <applied changes>

### Integration Check
- Flow compatibility: Pass | Fail
- Frontend-backend contract parity: Pass | Fail
- Docs updated: <files>

### Next Action
- /opsx:apply (tasks remaining) | /opsx-propose (new requirements) | /opsx:archive (all done)
```
