module video-streaming/backend/jams

go 1.21

require video-streaming/backend/shared v0.0.0

require (
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/segmentio/kafka-go v0.4.47 // indirect
)

replace video-streaming/backend/shared => ../shared
