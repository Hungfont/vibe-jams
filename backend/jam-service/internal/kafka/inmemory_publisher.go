package kafka

import "context"

// Record stores one in-memory published event for local development/testing.
type Record struct {
	Topic string
	Key   string
	Value []byte
}

// InMemoryPublisher captures records instead of sending to Kafka.
type InMemoryPublisher struct {
	Records []Record
}

// Publish appends a record to in-memory storage.
func (p *InMemoryPublisher) Publish(_ context.Context, topic string, key string, value []byte) error {
	p.Records = append(p.Records, Record{Topic: topic, Key: key, Value: value})
	return nil
}
