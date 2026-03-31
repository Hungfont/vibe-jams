## Context

`jam-service` create/end endpoints need consistent authorization behavior, but current auth integration is not standardized across the request path. The change spans `backend/jam-service`, `backend/auth-service`, and shared backend contracts, so implementation needs a clear interface contract, policy mapping, and common error handling semantics.

## Goals / Non-Goals

**Goals:**
- Provide a reusable auth entitlement guard for Jam create/end flows.
- Define a stable auth claim contract (`userId`, `plan`, `sessionState`) that `jam-service` can consume.
- Enforce premium entitlement policy with deterministic response mapping:
  - `401 unauthorized` for invalid/missing token or invalid session.
  - `403 premium_required` for authenticated users without premium plan.
- Add automated tests for policy correctness and cross-entrypoint consistency.

**Non-Goals:**
- Introducing a new identity provider or replacing token issuance strategy.
- Redesigning all API authorization middleware beyond Jam-related entrypoints.
- Changing business policy definitions for premium plans outside this story.

## Decisions

1. **Shared guard contract in backend shared layer**
   - Add/extend a shared interface and helper in `backend/shared` to parse and validate auth context.
   - Why: keeps auth semantics consistent across services and avoids duplicated claim parsing.
   - Alternative considered: implement separate guards in each service. Rejected due to drift risk and duplication.

2. **Auth-service remains source of truth for session/token validation**
   - `jam-service` delegates token/session validation to `auth-service` integration and consumes normalized claims.
   - Why: single source of truth for auth state reduces policy inconsistency.
   - Alternative considered: local token-only checks in Jam layer. Rejected because it cannot validate live session revocation state.

3. **Explicit claim contract**
   - Standardize required claim fields:
     - `userId` (string, required)
     - `plan` (enum/string, required for entitlement checks)
     - `sessionState` (valid/invalid, required)
   - Why: avoids implicit assumptions between service boundaries.
   - Alternative considered: pass through untyped auth payload. Rejected due to fragile integrations and runtime errors.

4. **Deterministic error mapping**
   - Guard returns normalized authorization outcomes mapped at API boundary:
     - unauthorized reasons -> HTTP 401 with `unauthorized`.
     - entitlement mismatch -> HTTP 403 with `premium_required`.
   - Why: prevents endpoint-specific error shape drift and improves client handling.
   - Alternative considered: service-specific error codes. Rejected due to inconsistent UX and contract complexity.

5. **Policy and contract tests as release gate**
   - Add tests for Jam create/end paths covering missing token, invalid token/session, non-premium, and premium success.
   - Why: this behavior is security-critical and must remain stable.
   - Alternative considered: manual verification only. Rejected because regression risk is high.

## Risks / Trade-offs

- **[Risk] Auth-service latency or outage affects Jam authorization path** -> **Mitigation:** use short timeout, explicit error classification, and clear fallback to `401 unauthorized` for validation failures.
- **[Risk] Claim contract drift between auth-service and jam-service** -> **Mitigation:** centralize contract type in shared package and add integration tests for deserialization/validation.
- **[Risk] Existing clients may rely on legacy error shape** -> **Mitigation:** document response contract and verify clients against `401 unauthorized`/`403 premium_required` before rollout.
- **[Trade-off] Additional integration check adds overhead per request** -> **Mitigation:** keep validation payload minimal and limit checks to protected Jam endpoints.
