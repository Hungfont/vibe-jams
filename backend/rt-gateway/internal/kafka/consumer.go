package kafka

import (
	"context"

	sharedkafka "video-streaming/backend/shared/kafka"
)

// Record is one consumed Kafka message.
type Record struct {
	Topic string
	Value []byte
}

// Consumer is the minimal fanout-consumer contract for rt-gateway.
type Consumer interface {
	Start(ctx context.Context, handler func(context.Context, Record) error) error
}

// NoopConsumer keeps process wiring active without broker dependency.
type NoopConsumer struct {
	delegate *sharedkafka.NoOpsConsumer
}

// NewNoopConsumer builds a consumer that exits only when context is done.
func NewNoopConsumer() *NoopConsumer {
	return &NoopConsumer{delegate: sharedkafka.NewNoOpsConsumer()}
}

// Start blocks until context cancellation.
func (c *NoopConsumer) Start(ctx context.Context, _ func(context.Context, Record) error) error {
	return c.delegate.Start(ctx)
}

// InMemoryConsumer is used by integration tests to inject records.
type InMemoryConsumer struct {
	recordCh chan Record
}

// NewInMemoryConsumer creates an in-memory consumer.
func NewInMemoryConsumer(buffer int) *InMemoryConsumer {
	if buffer <= 0 {
		buffer = 1
	}
	return &InMemoryConsumer{recordCh: make(chan Record, buffer)}
}

// Publish injects one record for Start() loop consumption.
func (c *InMemoryConsumer) Publish(record Record) bool {
	select {
	case c.recordCh <- record:
		return true
	default:
		return false
	}
}

// Start processes records until context cancellation.
func (c *InMemoryConsumer) Start(ctx context.Context, handler func(context.Context, Record) error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case record := <-c.recordCh:
			if err := handler(ctx, record); err != nil {
				return err
			}
		}
	}
}
