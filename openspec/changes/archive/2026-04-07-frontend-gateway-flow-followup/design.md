## Context

The backend change `api-gateway-authn-authz` is already completed and enforces gateway-based auth verification and BFF ingress. Frontend App Router API handlers currently preserve browser boundary (`/api/**`) but still use direct upstream service selection for auth/BFF flows. This causes contract drift versus gateway-first architecture and sequence documentation.

## Goals / Non-Goals

**Goals:**
- Align frontend auth and BFF orchestration upstream calls with api-gateway ingress.
- Preserve existing browser API surface and envelope contracts.
- Keep frontend/backend sequence documentation and runbook test flows aligned with actual behavior.
- Add regression tests for upstream routing decisions.

**Non-Goals:**
- Re-architect jam/create/join/playback service routing beyond current spec scope.
- Change browser-visible route paths.
- Modify backend gateway behavior in this change.

## Decisions

1. Use a dedicated gateway base URL in frontend config.
- Decision: add/consume `API_GATEWAY_URL` and route auth+BFF handlers via gateway.
- Rationale: explicit config avoids overloading auth/bff service env keys and keeps intent clear.
- Alternative considered: remap existing `AUTH_SERVICE_URL`/`API_SERVICE_URL` to gateway; rejected due to ambiguous semantics and migration risk.

2. Keep frontend browser routes unchanged.
- Decision: preserve `/api/auth/*` and `/api/bff/jam/{jamId}/orchestration` contracts.
- Rationale: avoids client breakage and allows backend topology change without UI contract churn.

3. Update docs in same change.
- Decision: update sequence doc plus runbook test flow after code updates.
- Rationale: repository guardrails require behavior and docs to remain synchronized.

4. Validate via focused route tests.
- Decision: extend/adjust frontend route tests that assert backend service selection and upstream paths.
- Rationale: fastest way to prevent regression in proxy routing behavior.

## Risks / Trade-offs

- [Risk] Existing environments may not set `API_GATEWAY_URL`.
  - Mitigation: define deterministic localhost default and keep backward-safe behavior for unrelated services.
- [Risk] Upstream auth errors may map differently through gateway hop.
  - Mitigation: keep existing frontend envelope normalization and verify with route tests.
- [Risk] Sequence doc drift.
  - Mitigation: update `docs/frontend-backend-sequence.md` in same change and add runbook flow.

## Migration Plan

1. Add gateway config entry and wire auth/BFF routes to use it.
2. Update/extend route tests for new upstream target assumptions.
3. Run frontend tests.
4. Update `docs/frontend-backend-sequence.md` and `docs/runbooks/run.md` with the new frontend execution flow.
5. Verify OpenSpec tasks and mark completion in apply phase.

## Open Questions

- Should future follow-up move jam/playback/catalog frontend server calls behind gateway as well, or keep current topology until a dedicated gateway capability is proposed?
