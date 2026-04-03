# Catalog Service Requirement File

## Dependency rank
- Rank: 2
- Dependency type: foundational
- Depends on: none

## Backend APIs used
- GET /internal/v1/catalog/tracks/{trackId}
- GET /healthz

## Frontend requirements
1. Queue add and playback commands must handle track_not_found deterministically.
2. Queue add and playback commands must handle track_unavailable deterministically.
3. Playable track metadata should enrich queue and now-playing display.
4. Lookup failures must not mutate local queue or playback state.

## Processing flow
1. User submits track action (add or playback target).
2. Frontend API route calls backend command path that includes catalog validation.
3. Response is normalized to success or deterministic error.
4. UI updates queue/player only on successful backend acceptance.
5. Error states keep existing room state unchanged and visible.

## Spotify-like shadcn component mapping
- Input + Button: quick add by track id in MVP flow.
- Card: track lookup result block with title and artist.
- Alert: track_not_found and track_unavailable explanation.
- Toast: lightweight failure notice without clearing queue view.

## Frontend router and page mapping

### App pages
- /jam/[jamId] -> frontend/src/app/jam/[jamId]/page.tsx

### App API routes
- GET /api/catalog/tracks/[trackId] -> frontend/src/app/api/catalog/tracks/[trackId]/route.ts -> GET /internal/v1/catalog/tracks/{trackId}

### Router flow
1. User inputs track id from /jam/[jamId].
2. UI calls /api/catalog/tracks/[trackId] for preview or precheck metadata.
3. Route maps track_not_found or track_unavailable and returns deterministic envelope.
4. Queue and playback actions only continue when response is playable.

## Success criteria
1. Same track errors behave identically in queue and playback surfaces.
2. Existing room state remains stable on lookup rejection.
3. Error text is actionable and maps to backend codes.
