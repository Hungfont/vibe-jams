package kafka

import (
	"context"
	"fmt"
	"strings"

	kafkago "github.com/segmentio/kafka-go"
)

// KafkaPublisher sends records to Kafka topics.
type KafkaPublisher struct {
	writer *kafkago.Writer
}

// NewKafkaPublisher builds a Kafka writer publisher.
func NewKafkaPublisher(bootstrapServers string) (*KafkaPublisher, error) {
	brokers := parseBrokers(bootstrapServers)
	if len(brokers) == 0 {
		return nil, fmt.Errorf("KAFKA_BOOTSTRAP_SERVERS is required")
	}

	return &KafkaPublisher{
		writer: &kafkago.Writer{
			Addr:         kafkago.TCP(brokers...),
			RequiredAcks: kafkago.RequireAll,
			Balancer:     &kafkago.Hash{},
		},
	}, nil
}

// Publish writes a record with topic/key/value.
func (p *KafkaPublisher) Publish(ctx context.Context, topic string, key string, value []byte) error {
	if strings.TrimSpace(topic) == "" {
		return fmt.Errorf("topic is required")
	}
	msg := kafkago.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: value,
	}
	return p.writer.WriteMessages(ctx, msg)
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
