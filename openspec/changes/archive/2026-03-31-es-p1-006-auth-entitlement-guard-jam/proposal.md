## Why

Jam create/end flows currently depend on auth integration points that are not fully standardized, which creates authorization gaps and inconsistent error behavior. We need a stable contract between `jam-service` and `auth-service` to enforce entitlement policy and unblock secure MVP delivery.

## What Changes

- Add a shared auth entitlement guard used by Jam create and end endpoints.
- Validate bearer tokens and normalize claims required by Jam flows (`userId`, `plan`, and token/session validity).
- Enforce premium-only authorization checks for protected Jam operations.
- Standardize authorization error mapping across entrypoints:
  - `401 unauthorized` for invalid, expired, or missing auth context.
  - `403 premium_required` for valid users without required premium entitlement.
- Add automated test coverage for premium/non-premium policy behavior and error response consistency.

## Capabilities

### New Capabilities
- `jam-auth-entitlement-guard`: Guard and middleware behavior for Jam create/end token validation and entitlement checks.
- `auth-claim-contract`: Stable claim payload contract consumed by Jam flows from `auth-service`.

### Modified Capabilities
- None.

## Impact

- Affected services: `backend/jam-service`, `backend/auth-service`, and shared auth/middleware contracts under `backend/shared`.
- API behavior changes: Jam authorization responses are normalized for 401/403 policy outcomes.
- Dependencies: auth token/session validation integration and shared middleware consumption in Jam API layer.
- Quality impact: adds authorization and policy tests to prevent regression in entitlement behavior.
