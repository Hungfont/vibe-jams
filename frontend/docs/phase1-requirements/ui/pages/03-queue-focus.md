# UI Requirement: Queue Focus

## URL
- /jam/[jamId]?view=queue

## Objective
1. Add remove reorder queue with deterministic consistency.

## Required UI components (shadcn)
1. ScrollArea
2. Card
3. DropdownMenu
4. Input
5. Button
6. Alert
7. Toast

## API callers
1. GET /api/jam/[jamId]/queue/snapshot -> GET /api/v1/jams/{jamId}/queue/snapshot
2. POST /api/jam/[jamId]/queue/add -> POST /api/v1/jams/{jamId}/queue/add
3. POST /api/jam/[jamId]/queue/remove -> POST /api/v1/jams/{jamId}/queue/remove
4. POST /api/jam/[jamId]/queue/reorder -> POST /api/v1/jams/{jamId}/queue/reorder
5. GET /api/catalog/tracks/[trackId] -> GET /internal/v1/catalog/tracks/{trackId}

## Requirements
1. Every queue mutation must honor queueVersion behavior.
2. version_conflict must trigger snapshot refresh and retry option.
3. Add flow must remain retry-safe under idempotent behavior.
4. track_not_found and track_unavailable must show explicit feedback.

## Processing flow
1. Queue snapshot renders current projection and version.
2. Mutation request is sent through frontend API route.
3. Success updates queue state.
4. Conflict triggers snapshot refetch then optional retry.
5. Track errors keep existing queue state unchanged.
