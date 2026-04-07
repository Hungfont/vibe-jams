---
name: /fullstack
id: fullstack
category: Workflow
description: Review OpenSpec coverage, update requirements/proposals, then implement backend and frontend with integration validation.
---

Run a fullstack OpenSpec-driven workflow from requirement review to integrated implementation.

**Input**: Optionally provide a change name or a short feature request after `/fullstack`.

**IMPORTANT**: Do NOT guess or auto-select a change. Always let the user choose.

**Primary path (always)**: `/opsx-propose` -> approval gate -> `/opsx:apply` (then `/opsx:archive` when complete).

**Purpose**

- Review feature requirements against current OpenSpec artifacts.
- Update artifacts when requirements are only additive.
- Automatically create or refresh OpenSpec change artifacts when requirements/features do not exist yet.
- Require explicit user approval after propose and before apply.
- Implement backend and frontend with explicit agent boundaries, in sequence by default and in parallel when safe.
- Verify integrated behavior still follows current documented flows.

## Agent Context Setup (Required)

1. Start in Agent (default) context for OpenSpec orchestration.
2. Load relevant OpenSpec skills first.
3. When implementation starts:
   - Use **Backend Engineer** for `backend/**` changes.
   - Use **Frontend Engineer** for `frontend/**` changes.
4. Do not mix backend and frontend implementation in one pass unless explicitly required by the task.

Announce active context and loaded sources before each phase.

## Workflow

1. **Resolve target change and requirement context**

   - If user provided a change name, use it.
   - Otherwise run:
     ```bash
     openspec list --json
     ```
   - If exactly one relevant active change exists, use it.
   - If multiple are possible, ask user to select.

2. **Review OpenSpec coverage for the requested feature**

   - Inspect status:
     ```bash
     openspec status --change "<name>" --json
     ```
   - Read relevant artifacts (proposal, design, specs, tasks).
   - Classify outcome:
     - **A. Additive requirement only**: existing feature exists, only extra acceptance/constraint is needed.
     - **B. Existing feature gap**: feature exists but required behavior is missing in artifacts.
     - **C. Missing feature**: no relevant requirement/feature exists in OpenSpec.

3. **Decision routing (proposal-first)**
   - Hard gate for completed changes:
     - Must run `/opsx-propose <new-change-name>` to create a new change and rebuild proposal, design, and tasks.

    - For **A** (additive requirement only):
       - Do not edit artifacts directly.
       - Run `/opsx-propose <name>` to refresh proposal/design/tasks for the requested update.
       - Then move to the mandatory user concern and approval gate before `/opsx:apply <name>`.
    - For **B** (existing feature gap):
       - Prefer `/opsx-propose <name>` first to refresh proposal/design/tasks against the new requirement.
       - Then move to the mandatory user concern and approval gate before `/opsx:apply <name>`.
    - For **C** (missing feature):
       - Prefer `/opsx-propose <new-change-name>` as the default path.
       - After artifacts are ready, move to the mandatory user concern and approval gate before `/opsx:apply <new-change-name>`.

4. **Update or create OpenSpec artifacts automatically**

    - For **A**, **B**, or updates on a completed change:
       - Use `/opsx-propose` to create or refresh change artifacts.
       - Rebuild proposal, design, and tasks before entering apply flow.
   - For **C**:
          - Prefer `/opsx-propose` for one-step artifact generation.
       - If slash-command path is unavailable or fails, use OpenSpec CLI fallback:
       ```bash
       openspec new change "<new-change-name>"
       openspec status --change "<new-change-name>" --json
       openspec instructions proposal --change "<new-change-name>" --json
       openspec instructions design --change "<new-change-name>" --json
       openspec instructions tasks --change "<new-change-name>" --json
       ```

5. **User concern and approval gate (mandatory before apply)**

      - Immediately after `/opsx-propose` completes, request user decision for the next step before any apply action.
      - This gate is a hard stop: no `/opsx:apply` execution is allowed until approval is explicitly captured.

      - Present a concise summary of proposed deltas before implementation:
         - OpenSpec artifacts to be updated
         - Expected backend changes
         - Expected frontend/integration changes
         - Potential contract or flow impact
      - Use the `vscode_askQuestions` tool to collect explicit decision:
         - Approve and continue to `/opsx:apply <name>`
         - I have concerns (stay in proposal phase)
      - Treat any response that is not explicit approval as not approved.
      - If user has concerns:
         - Stay in proposal phase and auto-suggest refinement options:
            1. Tighten or rewrite requirements in proposal
            2. Adjust acceptance criteria in specs
            3. Update design constraints and boundaries
            4. Split scope into follow-up change
            5. Keep current scope but document risks/assumptions
         - Allow custom user input in addition to predefined options.
         - Apply chosen refinements to proposal/design/spec/tasks.
         - Re-run concern check until explicit user approval is received.
      - If not approved, stop the apply path and return next action as proposal refinement.
      - Do not run `/opsx:apply` without explicit user approval.

6. **Prepare implementation plan from tasks**

      - Preconditions:
         - `/opsx-propose` is complete for the target change.
         - User approval status is explicitly `Approved`.
      - If approval status is not `Approved`, do not continue to apply.

   - Continue only on the primary path after approval: `/opsx:apply <name>`.
    - If slash-command path is unavailable or requires lower-level recovery, use apply instructions directly:
     ```bash
     openspec instructions apply --change "<name>" --json
     ```
    - Apply option visibility rule for session/progress:
       - In this case, next valid path for changes is `/opsx-propose`.
   - Confirm task order and identify impacted domains (`backend/**`, `frontend/**`).

7. **Implement backend and frontend execution strategy**

   - Default mode: backend-first, then frontend.
   - Parallel mode is allowed only when all conditions are true:
     - Proposed scope impacts both `backend/**` and `frontend/**`.
     - Tasks are independent enough to avoid ordering deadlock.
     - Contract decisions are already fixed in proposal/spec/design.
     - User has approved continuing to apply.
   - If parallel mode is selected, run domain agents in parallel while preserving strict folder boundaries.
   - Reconcile outputs before integration verification.

8. **Implement backend tasks (if backend is impacted)**

   - Switch to **Backend Engineer**.
   - Implement only backend tasks.
   - Keep contracts stable unless OpenSpec explicitly changed contracts.
   - Run backend validation/tests for touched services.

9. **Implement frontend tasks (if frontend is impacted)**

   - Switch to **Frontend Engineer**.
   - Implement only frontend tasks.
   - If backend contract changed, update frontend integration accordingly.
   - Keep frontend-to-backend orchestration aligned with `docs/frontend-backend-sequence.md`.
   - Run frontend validation/tests.

10. **Integration and flow verification**

   - Validate end-to-end behavior for changed flows.
   - Ensure current flow compatibility after integration.
   - Verify frontend-backend integration explicitly:
     - Frontend request/response types and adapters match backend contract.
     - Frontend no longer sends removed/deprecated fields.
     - Frontend handling for backend success/error envelope stays consistent.
   - If flow changed intentionally, update documentation in the same change:
     - `docs/frontend-backend-sequence.md`
     - `docs/runbooks/run.md`

11. **Finalize OpenSpec task tracking**

   - Mark completed tasks in the change tasks file (`- [ ]` -> `- [x]`).
   - Summarize what was updated in OpenSpec and what was implemented in code.
   - If all tasks are done, suggest `/opsx:archive`.

## Guardrails

- Never skip OpenSpec update when requirement changes are discovered.
- Primary path is always: `/opsx-propose` -> approval gate -> `/opsx:apply`.
- After all tasks are complete, suggest `/opsx:archive`.
- Use raw OpenSpec CLI as fallback only when slash-command path is unavailable or blocked.
- Mandatory approval gate: explicit user approval is required before `/opsx:apply`, immediately after propose output is presented.
- Hard stop rule: if approval is pending or concerns remain, continue proposal refinement and do not enter apply.
- Parallel backend/frontend implementation is allowed only after approval and only when dependencies are safe for concurrent execution.
- If user raises concerns, remain in proposal phase and refine artifacts before any implementation.
- If requirements are unclear, ask for clarification before coding.
- Keep changes minimal, scoped, and domain-correct.
- Preserve backend/frontend boundaries by agent.
- Ensure runbook and sequence docs reflect real behavior when changed.
- Verify frontend-backend integration alignment before completion.
- Stop only when work is complete or genuinely blocked.

## Output Template

```md
## Fullstack Execution Summary

**Change:** <name>
**OpenSpec Action:** Updated existing artifacts | Created new change
**Coverage Result:** A | B | C
**Primary Path:** /opsx-propose -> approval gate -> /opsx:apply (then /opsx:archive when complete)
**Approval Status:** Approved | Pending concerns
**Approval Checkpoint:** Requested after propose | Not requested (invalid)

### OpenSpec Updates
- <artifact updates>

### Concerns and Resolution
- User concerns: <none or summary>
- Selected option(s): <1-5 or custom>
- Proposal-phase refinements: <applied changes>

### Backend Implementation (Backend Engineer)
- <changes>
- Validation: <results>

### Frontend Implementation (Frontend Engineer)
- <changes>
- Validation: <results>

### Integration Check
- Flow compatibility: Pass | Fail
- Frontend-backend contract parity: Pass | Fail
- Docs updated: <files>

### Task Progress
- Completed: <n>/<m>
- Next action: /opsx:apply (when remaining > 0) | /opsx-propose (when all_done and updates are requested) | /opsx:archive (when all_done and no further updates) | Follow-up clarification
```