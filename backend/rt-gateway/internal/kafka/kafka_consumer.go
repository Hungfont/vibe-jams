package kafka

import (
	"context"
	"fmt"
	"strings"

	kafkago "github.com/segmentio/kafka-go"
)

// KafkaConsumer reads records from Kafka topics using a consumer group.
type KafkaConsumer struct {
	readers []*kafkago.Reader
}

// NewKafkaConsumer builds one reader per topic.
func NewKafkaConsumer(bootstrapServers string, groupID string, topics []string) (*KafkaConsumer, error) {
	brokers := parseBrokers(bootstrapServers)
	if len(brokers) == 0 {
		return nil, fmt.Errorf("KAFKA_BOOTSTRAP_SERVERS is required")
	}
	if strings.TrimSpace(groupID) == "" {
		return nil, fmt.Errorf("KAFKA_CONSUMER_GROUP is required")
	}
	if len(topics) == 0 {
		return nil, fmt.Errorf("at least one Kafka topic is required")
	}

	readers := make([]*kafkago.Reader, 0, len(topics))
	for _, topic := range topics {
		topic = strings.TrimSpace(topic)
		if topic == "" {
			continue
		}
		readers = append(readers, kafkago.NewReader(kafkago.ReaderConfig{
			Brokers:  brokers,
			GroupID:  groupID,
			Topic:    topic,
			MinBytes: 1,
			MaxBytes: 10e6,
		}))
	}
	if len(readers) == 0 {
		return nil, fmt.Errorf("no Kafka topics configured")
	}
	return &KafkaConsumer{readers: readers}, nil
}

// Start begins consuming records and forwarding them to handler.
func (c *KafkaConsumer) Start(ctx context.Context, handler func(context.Context, Record) error) error {
	errCh := make(chan error, len(c.readers))

	for _, reader := range c.readers {
		r := reader
		go func() {
			for {
				msg, err := r.FetchMessage(ctx)
				if err != nil {
					if ctx.Err() != nil {
						return
					}
					errCh <- fmt.Errorf("fetch Kafka message: %w", err)
					return
				}

				if err := handler(ctx, Record{Topic: msg.Topic, Value: msg.Value}); err != nil {
					errCh <- err
					return
				}

				if err := r.CommitMessages(ctx, msg); err != nil {
					if ctx.Err() != nil {
						return
					}
					errCh <- fmt.Errorf("commit Kafka message: %w", err)
					return
				}
			}
		}()
	}

	select {
	case <-ctx.Done():
		c.close()
		return ctx.Err()
	case err := <-errCh:
		c.close()
		return err
	}
}

func (c *KafkaConsumer) close() {
	for _, reader := range c.readers {
		_ = reader.Close()
	}
}

func parseBrokers(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}
