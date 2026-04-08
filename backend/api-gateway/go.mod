module video-streaming/backend/api-gateway

go 1.21

replace video-streaming/backend/shared => ../shared

require (
	github.com/golang-jwt/jwt/v5 v5.3.1
	video-streaming/backend/shared v0.0.0-00010101000000-000000000000
)
