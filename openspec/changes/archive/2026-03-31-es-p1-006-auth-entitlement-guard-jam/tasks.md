## 1. Shared Auth Contract

- [x] 1.1 Define shared auth claim type in `backend/shared` with required fields (`userId`, `plan`, `sessionState`)
- [x] 1.2 Implement/extend claim validation helper to fail fast on missing required fields
- [x] 1.3 Add unit tests for claim parsing and validation edge cases (missing fields, invalid session state)

## 2. Auth-Service Validation Integration

- [x] 2.1 Implement token/session validation endpoint or integration path used by Jam authorization
- [x] 2.2 Normalize auth-service validation output to shared claim contract and error categories
- [x] 2.3 Add auth-service tests for valid token, expired/invalid token, and revoked session handling

## 3. Jam-Service Entitlement Guard

- [x] 3.1 Add shared entitlement guard middleware/handler integration for Jam create endpoint
- [x] 3.2 Add shared entitlement guard middleware/handler integration for Jam end endpoint
- [x] 3.3 Implement deterministic entitlement policy mapping from claim `plan` to premium/non-premium outcomes

## 4. Error Mapping and API Consistency

- [x] 4.1 Standardize unauthorized outcomes to HTTP `401` with `unauthorized` across Jam entrypoints
- [x] 4.2 Standardize insufficient entitlement outcomes to HTTP `403` with `premium_required` across Jam entrypoints
- [x] 4.3 Add tests validating consistent error envelope/code mapping across create/end flows

## 5. End-to-End Verification and Rollout Safety

- [x] 5.1 Add integration tests for Jam create/end matrix (missing token, invalid token, invalid session, non-premium, premium)
- [x] 5.2 Document the auth claim contract and 401/403 behavior in backend service docs/runbook notes
- [x] 5.3 Execute service-level test suites and confirm no regression in existing Jam flow behavior
