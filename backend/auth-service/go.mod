module video-streaming/backend/auth-service

go 1.21

require video-streaming/backend/shared v0.0.0

require (
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/lib/pq v1.10.9
)

replace video-streaming/backend/shared => ../shared
