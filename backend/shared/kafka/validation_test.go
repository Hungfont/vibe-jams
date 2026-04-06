package kafka

import "testing"

func TestValidateTopicBaseline(t *testing.T) {
	actual := []ActualTopicConfig{
		{Name: TopicJamSession, Partitions: 12, RetentionHours: 168, ProducerPrincipals: []string{"svc-jam-service"}, ConsumerPrincipals: []string{}},
		{Name: TopicJamQueue, Partitions: 24, RetentionHours: 168, ProducerPrincipals: []string{"svc-jam-service"}, ConsumerPrincipals: []string{"svc-rt-gateway"}},
		{Name: TopicJamPlayback, Partitions: 12, RetentionHours: 168, ProducerPrincipals: []string{"svc-playback-service"}, ConsumerPrincipals: []string{"svc-rt-gateway"}},
		{Name: TopicJamModeration, Partitions: 12, RetentionHours: 168, ProducerPrincipals: []string{"svc-jam-service"}, ConsumerPrincipals: []string{"svc-rt-gateway"}},
		{Name: TopicAnalyticsUser, Partitions: 24, RetentionHours: 336, ProducerPrincipals: []string{"svc-api-service"}, ConsumerPrincipals: []string{}},
	}

	expectedProducers := map[string][]string{
		TopicJamSession:    {"svc-jam-service"},
		TopicJamQueue:      {"svc-jam-service"},
		TopicJamPlayback:   {"svc-playback-service"},
		TopicJamModeration: {"svc-jam-service"},
		TopicAnalyticsUser: {"svc-api-service"},
	}
	expectedConsumers := map[string][]string{
		TopicJamQueue:      {"svc-rt-gateway"},
		TopicJamPlayback:   {"svc-rt-gateway"},
		TopicJamModeration: {"svc-rt-gateway"},
	}

	if err := ValidateTopicBaseline(actual, expectedProducers, expectedConsumers); err != nil {
		t.Fatalf("ValidateTopicBaseline() error = %v", err)
	}
}
