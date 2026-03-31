package kafka

import (
	"fmt"
	"slices"
	"time"
)

// ActualTopicConfig represents discovered environment topic settings.
type ActualTopicConfig struct {
	Name               string   `json:"name"`
	Partitions         int      `json:"partitions"`
	RetentionHours     int      `json:"retentionHours"`
	ProducerPrincipals []string `json:"producerPrincipals"`
	ConsumerPrincipals []string `json:"consumerPrincipals"`
}

// ValidateTopicBaseline verifies expected topic settings and ACL principals.
func ValidateTopicBaseline(actual []ActualTopicConfig, expectedProducers map[string][]string, expectedConsumers map[string][]string) error {
	byName := make(map[string]ActualTopicConfig, len(actual))
	for _, cfg := range actual {
		byName[cfg.Name] = cfg
	}

	for _, expected := range Phase1TopicConfigs {
		got, ok := byName[expected.Name]
		if !ok {
			return fmt.Errorf("missing topic %s", expected.Name)
		}
		if got.Partitions != expected.Partitions {
			return fmt.Errorf("topic %s partitions mismatch: got %d want %d", expected.Name, got.Partitions, expected.Partitions)
		}

		wantRetentionHours := int(expected.Retention / time.Hour)
		if got.RetentionHours != wantRetentionHours {
			return fmt.Errorf("topic %s retention mismatch: got %dh want %dh", expected.Name, got.RetentionHours, wantRetentionHours)
		}

		for _, principal := range expectedProducers[expected.Name] {
			if !slices.Contains(got.ProducerPrincipals, principal) {
				return fmt.Errorf("topic %s missing producer ACL for %s", expected.Name, principal)
			}
		}
		for _, principal := range expectedConsumers[expected.Name] {
			if !slices.Contains(got.ConsumerPrincipals, principal) {
				return fmt.Errorf("topic %s missing consumer ACL for %s", expected.Name, principal)
			}
		}
	}

	return nil
}
