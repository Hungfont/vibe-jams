module video-streaming/backend/rt-gateway

go 1.21

require (
	github.com/gorilla/websocket v1.5.3
	video-streaming/backend/shared v0.0.0
)

replace video-streaming/backend/shared => ../shared
