package kafka

import (
	"context"
	"errors"
	"testing"
)

func TestNoopConsumerStartReturnsContextCanceled(t *testing.T) {
	t.Parallel()

	consumer := NewNoopConsumer()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := consumer.Start(ctx, func(context.Context, Record) error {
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Start() error = %v, want context canceled", err)
	}
}
