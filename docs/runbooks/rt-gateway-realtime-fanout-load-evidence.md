# RT-Gateway Realtime Fanout Load Evidence (ES-P1-003)

## Scenario

- Test: `TestLoadScenario_FanoutP95LatencyUnderTarget`
- Module: `backend/rt-gateway/internal/fanout`
- Command:

```bash
go test ./backend/rt-gateway/internal/fanout -run TestLoadScenario_FanoutP95LatencyUnderTarget -count=1
```

## What It Validates

- 1,000 ordered events fan out for a single session stream.
- Per-event processing records fanout latency metrics.
- Test fails if p95 fanout latency exceeds 50ms.

## Acceptance

- Pass criteria: test exits successfully with p95 <= 50ms.
- Evidence capture: retain CI test log artifact for release sign-off.
