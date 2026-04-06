## 1. Jam-Service Moderation Commands

- [x] 1.1 Add moderation domain models and repository state transitions for `mute` and `kick`
- [x] 1.2 Implement host-only moderation service methods and deterministic error mapping
- [x] 1.3 Add jam-service HTTP moderation handlers and route parsing for `/api/v1/jams/{jamId}/moderation/{action}`
- [x] 1.4 Enforce blocked queue actions for muted or kicked participants

## 2. Kafka and Fanout Integration

- [x] 2.1 Add shared Kafka topic constant/config for `jam.moderation.events` and update topic validation fixtures
- [x] 2.2 Extend jam-service event producer to publish moderation audit envelopes
- [x] 2.3 Extend rt-gateway config/consumer subscriptions to include moderation topic
- [x] 2.4 Add rt-gateway moderation consumer hook interface for abuse heuristics and wire default no-op

## 3. Frontend Integration

- [x] 3.1 Add frontend API routes for moderation mute and kick commands
- [x] 3.2 Update jam room participant model and UI to show moderation state and host moderation controls
- [x] 3.3 Ensure room updates reflect moderation changes via realtime or snapshot refresh path

## 4. Documentation and Validation

- [x] 4.1 Update `docs/frontend-backend-sequence.md` moderation command and moderation event flow
- [x] 4.2 Update `docs/runbooks/run.md` with moderation execution flow and deterministic outcomes
- [x] 4.3 Add backend integration tests for moderation command behavior and blocked-action enforcement
- [x] 4.4 Add gateway/frontend integration tests for moderation event fanout and room visibility