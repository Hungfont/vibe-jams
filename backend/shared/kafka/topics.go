package kafka

import "time"

const (
	TopicJamSession    = "jam.session.events"
	TopicJamQueue      = "jam.queue.events"
	TopicJamPlayback   = "jam.playback.events"
	TopicJamModeration = "jam.moderation.events"
	TopicAnalyticsUser = "analytics.user.actions"
)

// TopicConfig defines Kafka topic policy values used during provisioning checks.
type TopicConfig struct {
	Name       string
	Partitions int
	Retention  time.Duration
}

// Phase1TopicConfigs maps the LLD baseline topic settings for MVP.
var Phase1TopicConfigs = []TopicConfig{
	{Name: TopicJamSession, Partitions: 12, Retention: 7 * 24 * time.Hour},
	{Name: TopicJamQueue, Partitions: 24, Retention: 7 * 24 * time.Hour},
	{Name: TopicJamPlayback, Partitions: 12, Retention: 7 * 24 * time.Hour},
	{Name: TopicJamModeration, Partitions: 12, Retention: 7 * 24 * time.Hour},
	{Name: TopicAnalyticsUser, Partitions: 24, Retention: 14 * 24 * time.Hour},
}
