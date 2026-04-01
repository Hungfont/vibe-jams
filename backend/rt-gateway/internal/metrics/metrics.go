package metrics

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Snapshot is a serializable read view of fanout metrics.
type Snapshot struct {
	FanoutCount          int64   `json:"fanoutCount"`
	DuplicateCount       int64   `json:"duplicateCount"`
	GapDetectedCount     int64   `json:"gapDetectedCount"`
	RecoverySuccessCount int64   `json:"recoverySuccessCount"`
	RecoveryFailureCount int64   `json:"recoveryFailureCount"`
	SlowConsumerCount    int64   `json:"slowConsumerCount"`
	P95FanoutLatencyMS   float64 `json:"p95FanoutLatencyMs"`
	P95ConsumerLagMS     float64 `json:"p95ConsumerLagMs"`
	P95SnapshotLatencyMS float64 `json:"p95SnapshotLatencyMs"`
}

// Registry tracks fanout counters and latency distributions.
type Registry struct {
	fanoutCount          atomic.Int64
	duplicateCount       atomic.Int64
	gapDetectedCount     atomic.Int64
	recoverySuccessCount atomic.Int64
	recoveryFailureCount atomic.Int64
	slowConsumerCount    atomic.Int64

	mu              sync.Mutex
	fanoutLatencies []time.Duration
	consumerLag     []time.Duration
	snapshotLatency []time.Duration
}

// NewRegistry creates an in-memory metrics registry.
func NewRegistry() *Registry {
	return &Registry{
		fanoutLatencies: make([]time.Duration, 0, 1024),
		consumerLag:     make([]time.Duration, 0, 1024),
		snapshotLatency: make([]time.Duration, 0, 512),
	}
}

func (r *Registry) IncFanout() {
	r.fanoutCount.Add(1)
}

func (r *Registry) IncDuplicate() {
	r.duplicateCount.Add(1)
}

func (r *Registry) IncGapDetected() {
	r.gapDetectedCount.Add(1)
}

func (r *Registry) IncRecoverySuccess() {
	r.recoverySuccessCount.Add(1)
}

func (r *Registry) IncRecoveryFailure() {
	r.recoveryFailureCount.Add(1)
}

func (r *Registry) IncSlowConsumer() {
	r.slowConsumerCount.Add(1)
}

func (r *Registry) ObserveFanoutLatency(duration time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fanoutLatencies = append(r.fanoutLatencies, duration)
}

func (r *Registry) ObserveConsumerLag(duration time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.consumerLag = append(r.consumerLag, duration)
}

func (r *Registry) ObserveSnapshotLatency(duration time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.snapshotLatency = append(r.snapshotLatency, duration)
}

// Snapshot returns current counters and p95 latency values.
func (r *Registry) Snapshot() Snapshot {
	r.mu.Lock()
	fanout := cloneDurations(r.fanoutLatencies)
	lag := cloneDurations(r.consumerLag)
	snapshotLatency := cloneDurations(r.snapshotLatency)
	r.mu.Unlock()

	return Snapshot{
		FanoutCount:          r.fanoutCount.Load(),
		DuplicateCount:       r.duplicateCount.Load(),
		GapDetectedCount:     r.gapDetectedCount.Load(),
		RecoverySuccessCount: r.recoverySuccessCount.Load(),
		RecoveryFailureCount: r.recoveryFailureCount.Load(),
		SlowConsumerCount:    r.slowConsumerCount.Load(),
		P95FanoutLatencyMS:   durationP95Milliseconds(fanout),
		P95ConsumerLagMS:     durationP95Milliseconds(lag),
		P95SnapshotLatencyMS: durationP95Milliseconds(snapshotLatency),
	}
}

func cloneDurations(values []time.Duration) []time.Duration {
	out := make([]time.Duration, len(values))
	copy(out, values)
	return out
}

func durationP95Milliseconds(values []time.Duration) float64 {
	if len(values) == 0 {
		return 0
	}
	sort.Slice(values, func(i, j int) bool {
		return values[i] < values[j]
	})
	index := int(float64(len(values)-1) * 0.95)
	if index < 0 {
		index = 0
	}
	if index >= len(values) {
		index = len(values) - 1
	}
	return float64(values[index].Microseconds()) / 1000
}
