# UI Requirement: Participants and Role

## URL
- /jam/[jamId]?view=participants

## Objective
1. Show participant role and entitlement context clearly.
2. Drive permission states from role and plan.

## Required UI components (shadcn)
1. Avatar
2. Badge
3. Tooltip
4. Separator

## API callers
1. POST /api/bff/jam/[jamId]/orchestration -> POST /v1/bff/mvp/sessions/{sessionId}/orchestration
2. GET /api/jam/[jamId]/state -> GET /api/v1/jams/{jamId}/state
3. WebSocket /ws?sessionId={id}&lastSeenVersion={optional} -> rt-gateway

## Requirements
1. Participant list must show host and member roles.
2. Header must show premium and session status badges.
3. Controls must disable by role and entitlement policy.

## Processing flow
1. Page reads participant context from orchestration.
2. UI renders role and plan badges.
3. Realtime updates keep participant list synchronized.
4. Permission states update with each new session snapshot.
