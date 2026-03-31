## Jam/Auth entitlement contract

This runbook defines the auth claim and error contract used by `jam-service` and `auth-service`.

### Claim contract

`auth-service` endpoint: `POST /internal/v1/auth/validate`

Successful response (`200`) MUST include:

- `userId` (string, required)
- `plan` (string, required)
- `sessionState` (`valid` or `invalid`, required)

### Jam error mapping

For guarded Jam endpoints (`/api/v1/jams/create`, `/api/v1/jams/{jamId}/end`):

- `401 unauthorized`
  - missing bearer token
  - invalid/expired token
  - missing required claims
  - invalid session state
- `403 premium_required`
  - authenticated claim is valid but `plan` is non-premium

### Test tokens for local verification

- `token-premium-valid` -> premium user, session valid
- `token-free-valid` -> free user, session valid
- `token-premium-revoked` -> premium user, session invalid
