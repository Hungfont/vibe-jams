package kafka

import (
	"context"
	"errors"
	"testing"
)

func TestNoOpsProducerPublishAlwaysSucceeds(t *testing.T) {
	t.Parallel()

	producer := NewNoOpsProducer()
	if err := producer.Publish(context.Background(), "topic", "key", []byte("payload")); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
}

func TestNoOpsConsumerStartReturnsContextError(t *testing.T) {
	t.Parallel()

	consumer := NewNoOpsConsumer()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := consumer.Start(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Start() error = %v, want context canceled", err)
	}
}
