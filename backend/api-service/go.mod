module video-streaming/backend/api-service

go 1.21

require video-streaming/backend/shared v0.0.0

require github.com/golang-jwt/jwt/v5 v5.2.2 // indirect

replace video-streaming/backend/shared => ../shared
