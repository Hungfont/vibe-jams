# Service Requirement Ranking by Dependency (Phase 1)

## Ranking rules
1. Services that provide identity and validation contracts are first.
2. Services required by write paths are before orchestration and fanout optimization.
3. Aggregation and realtime experience layers are after core mutation paths.

## Ranked files
1. 01-auth-service-requirements.md
- Dependency level: foundational
- Depends on: none
- Required by: jam-service, playback-service, api-service-bff

2. 02-catalog-service-requirements.md
- Dependency level: foundational
- Depends on: none
- Required by: jam-service, playback-service, api-service-bff

3. 03-jam-service-requirements.md
- Dependency level: core state owner
- Depends on: auth-service, catalog-service
- Required by: playback-service conflict checks, rt-gateway recovery, api-service-bff orchestration

4. 04-playback-service-requirements.md
- Dependency level: core command path
- Depends on: auth-service, catalog-service, jam queue version model
- Required by: api-service-bff optional playback segment, rt-gateway playback fanout

5. 05-rt-gateway-requirements.md
- Dependency level: realtime sync layer
- Depends on: jam-service snapshot, queue and playback events
- Required by: room synchronization and reconnect recovery UX

6. 06-api-service-bff-requirements.md
- Dependency level: aggregation layer
- Depends on: auth-service, jam-service, catalog-service, playback-service
- Required by: first load room hydration and partial-dependency UX

## Delivery order for implementation
1. auth-service requirements
2. catalog-service requirements
3. jam-service requirements
4. playback-service requirements
5. rt-gateway requirements
6. api-service-bff requirements

## Router file priority by dependency
1. frontend/src/app/page.tsx
2. frontend/src/app/jam/[jamId]/page.tsx
3. frontend/src/app/api/auth/validate/route.ts
4. frontend/src/app/api/catalog/tracks/[trackId]/route.ts
5. frontend/src/app/api/jam/create/route.ts
6. frontend/src/app/api/jam/[jamId]/join/route.ts
7. frontend/src/app/api/jam/[jamId]/leave/route.ts
8. frontend/src/app/api/jam/[jamId]/end/route.ts
9. frontend/src/app/api/jam/[jamId]/state/route.ts
10. frontend/src/app/api/jam/[jamId]/queue/add/route.ts
11. frontend/src/app/api/jam/[jamId]/queue/remove/route.ts
12. frontend/src/app/api/jam/[jamId]/queue/reorder/route.ts
13. frontend/src/app/api/jam/[jamId]/queue/snapshot/route.ts
14. frontend/src/app/api/jam/[jamId]/playback/commands/route.ts
15. frontend/src/app/api/bff/jam/[jamId]/orchestration/route.ts
16. frontend/src/app/api/realtime/ws-config/route.ts

## UI page and URL design reference
- ../ui/01-page-url-api-design.md
