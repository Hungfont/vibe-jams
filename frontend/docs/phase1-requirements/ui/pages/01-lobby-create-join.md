# UI Requirement: Lobby Create and Join

## URL
- /

## Objective
1. Let user create or join Jam quickly.
2. Navigate to room immediately on success.

## Required UI components (shadcn)
1. Card
2. Tabs
3. Input
4. Button
5. Alert
6. Toast

## API callers
1. POST /api/auth/validate -> POST /internal/v1/auth/validate
2. POST /api/jam/create -> POST /api/v1/jams/create
3. POST /api/jam/[jamId]/join -> POST /api/v1/jams/{jamId}/join

## Requirements
1. Provide two actions: Create Jam and Join Jam.
2. Create must check entitlement before submit.
3. Join must validate jamId and show explicit errors.
4. Success must redirect to /jam/[jamId].

## Processing flow
1. User selects create or join tab.
2. UI submits to frontend API route.
3. API route validates auth and entitlement.
4. API route returns jamId or normalized error.
5. UI redirects to room or shows error feedback.
