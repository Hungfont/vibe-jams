# API Service BFF Requirement File

## Dependency rank
- Rank: 6
- Dependency type: aggregation layer
- Depends on: auth-service, jam-service, catalog-service, playback-service

## Backend APIs used
- POST /v1/bff/mvp/sessions/{sessionId}/orchestration
- GET /healthz

## Frontend requirements
1. Jam room first load should use one orchestration request.
2. auth and jam are required segments and must fail fast.
3. catalog and playback are optional segments and may degrade with partial=true.
4. UI must render core session and queue even when optional segments fail.
5. Dependency issues must be shown as actionable notices.

## Processing flow
1. Room page requests orchestration payload via frontend API route.
2. Frontend API route forwards auth context and session id.
3. BFF returns aggregated data with dependency statuses and optional issues.
4. UI renders required room data immediately.
5. Optional failures are displayed in non-blocking degraded panel and action-level disables.

## Spotify-like shadcn component mapping
- Alert: dependency degraded banner.
- Card: segmented room panels that remain interactive where possible.
- Tabs: split queue, participants, diagnostics in one room layout.
- Tooltip: explain disabled optional controls when dependency degraded.

## Frontend router and page mapping

### App pages
- /jam/[jamId] -> frontend/src/app/jam/[jamId]/page.tsx

### App API routes
- POST /api/bff/jam/[jamId]/orchestration -> frontend/src/app/api/bff/jam/[jamId]/orchestration/route.ts -> POST /v1/bff/mvp/sessions/{sessionId}/orchestration

### Router flow
1. /jam/[jamId] server-side load calls /api/bff/jam/[jamId]/orchestration.
2. Route forwards auth context and optional orchestration payload.
3. Response hydrates required room blocks first.
4. partial=true displays dependency issue panel without breaking core room state.
5. Optional controls are disabled only for degraded dependencies.

## Success criteria
1. One request hydrates initial room state.
2. partial=true does not break required room interaction.
3. Dependency failure feedback is explicit and actionable.
