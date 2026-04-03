# UI Requirement Index by Page (Phase 1)

## Shared principles
1. Browser calls only frontend API routes.
2. Frontend API routes map to backend APIs with normalized error envelope.
3. UI uses shadcn primitives and Tailwind utilities only.

## Page requirement files
1. 01-lobby-create-join.md
- URL: /

2. 02-jam-room-shell.md
- URL: /jam/[jamId]

3. 03-queue-focus.md
- URL: /jam/[jamId]?view=queue

4. 04-playback-focus.md
- URL: /jam/[jamId]?view=playback

5. 05-participants-role.md
- URL: /jam/[jamId]?view=participants

6. 06-diagnostics-degraded.md
- URL: /jam/[jamId]?view=diagnostics

## Recommended UI build order
1. /
2. /jam/[jamId]
3. /jam/[jamId]?view=queue
4. /jam/[jamId]?view=playback
5. /jam/[jamId]?view=participants
6. /jam/[jamId]?view=diagnostics
