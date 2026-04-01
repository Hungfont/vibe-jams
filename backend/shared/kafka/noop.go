package kafka

import "context"

// NoOpsProducer is a no-op publisher used for non-Kafka fallback paths.
type NoOpsProducer struct{}

// NewNoOpsProducer builds a no-op Kafka producer implementation.
func NewNoOpsProducer() *NoOpsProducer {
	return &NoOpsProducer{}
}

// Publish intentionally performs no operation and always succeeds.
func (p *NoOpsProducer) Publish(_ context.Context, _ string, _ string, _ []byte) error {
	return nil
}

// NoOpsConsumer is a no-op consumer used for non-Kafka fallback paths.
type NoOpsConsumer struct{}

// NewNoOpsConsumer builds a no-op Kafka consumer implementation.
func NewNoOpsConsumer() *NoOpsConsumer {
	return &NoOpsConsumer{}
}

// Start blocks until the context is canceled.
func (c *NoOpsConsumer) Start(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}
