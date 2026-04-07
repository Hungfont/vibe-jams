package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"video-streaming/backend/playback-service/internal/auth"
	"video-streaming/backend/playback-service/internal/config"
	"video-streaming/backend/playback-service/internal/handler"
	"video-streaming/backend/playback-service/internal/kafka"
	"video-streaming/backend/playback-service/internal/repository"
	"video-streaming/backend/playback-service/internal/service"
	sharedcatalog "video-streaming/backend/shared/catalog"
	sharedkafka "video-streaming/backend/shared/kafka"
)

const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Playback Service API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function() {
      SwaggerUIBundle({
        url: '/swagger/openapi.json',
        dom_id: '#swagger-ui'
      });
    };
  </script>
</body>
</html>`

const openAPISpec = `{
  "openapi": "3.0.3",
  "info": {
    "title": "Playback Service API",
    "version": "1.0.0"
  },
  "servers": [
    {
      "url": "http://localhost:8082"
    }
  ],
  "paths": {
    "/healthz": {
      "get": {
        "summary": "Health check",
        "responses": {
          "200": {
            "description": "Service is healthy"
          }
        }
      }
    },
    "/v1/jam/sessions/{sessionId}/playback/commands": {
      "post": {
        "summary": "Submit host playback command",
        "parameters": [
          {
            "name": "sessionId",
            "in": "path",
            "required": true,
            "schema": { "type": "string" }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/PlaybackCommandRequest" }
            }
          }
        },
        "responses": {
          "202": {
            "description": "Command accepted",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/CommandAcceptedResponse" }
              }
            }
          },
          "400": {
            "description": "Invalid request",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/ErrorBody" }
              }
            }
          },
          "401": {
            "description": "Unauthorized",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/ErrorBody" }
              }
            }
          },
          "403": {
            "description": "Host only",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/ErrorBody" }
              }
            }
          },
          "404": {
            "description": "Session not found",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/ErrorBody" }
              }
            }
          },
          "409": {
            "description": "Version conflict",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/ErrorBody" }
              }
            }
          },
          "500": {
            "description": "Internal server error",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/ErrorBody" }
              }
            }
          }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "PlaybackCommandRequest": {
        "type": "object",
        "properties": {
          "command": { "type": "string" },
          "trackId": { "type": "string" },
          "clientEventId": { "type": "string" },
          "expectedQueueVersion": { "type": "integer", "format": "int64" },
          "positionMs": { "type": "integer", "format": "int64" }
        },
        "required": ["command", "clientEventId", "expectedQueueVersion"]
      },
      "CommandAcceptedResponse": {
        "type": "object",
        "properties": {
      "accepted": { "type": "boolean" },
      "queueVersion": { "type": "integer", "format": "int64" },
      "playbackEpoch": { "type": "integer", "format": "int64" }
        },
    "required": ["accepted", "queueVersion", "playbackEpoch"]
      },
      "ErrorDetail": {
        "type": "object",
        "properties": {
          "code": { "type": "string" },
      "message": { "type": "string" },
      "retry": {
      "type": "object",
      "properties": {
        "currentQueueVersion": { "type": "integer", "format": "int64" },
        "playbackEpoch": { "type": "integer", "format": "int64" }
      },
      "required": ["currentQueueVersion"]
      }
        },
        "required": ["code", "message"]
      },
      "ErrorBody": {
        "type": "object",
        "properties": {
          "error": { "$ref": "#/components/schemas/ErrorDetail" }
        },
        "required": ["error"]
      }
    }
  }
}`

type healthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}

// NewRouter builds all playback-service HTTP routes.
func NewRouter(cfg config.Config) (http.Handler, error) {
	if err := cfg.ValidateRuntimePolicy(); err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	var repo *repository.RedisPlaybackRepository
	switch strings.ToLower(cfg.StateStoreBackend) {
	case "inmemory":
		repo = repository.NewRedisPlaybackRepository()
	case "redis", "postgres":
		durableRepo, err := repository.NewDurablePlaybackRepository(cfg.StateStorePath)
		if err != nil {
			return nil, err
		}
		repo = durableRepo
	default:
		return nil, fmt.Errorf("unsupported STATE_STORE_BACKEND: %s", cfg.StateStoreBackend)
	}

	// Local seed keeps a testable baseline session for manual smoke checks.
	_ = repo.SeedSession("jam-local", "host-local", 1)

	var publisher kafka.Publisher
	switch strings.ToLower(cfg.KafkaTransport) {
	case "kafka":
		kp, err := kafka.NewKafkaPublisher(cfg.KafkaBootstrapServers)
		if err != nil {
			return nil, err
		}
		publisher = kp
	case "inmemory":
		publisher = &kafka.InMemoryPublisher{}
	default:
		publisher = sharedkafka.NewNoOpsProducer()
		return nil, fmt.Errorf("unsupported KAFKA_TRANSPORT: %s", cfg.KafkaTransport)
	}

	producer := kafka.NewProducer(publisher)
	var catalogValidator sharedcatalog.Validator
	if cfg.EnableCatalogValidation {
		catalogValidator = sharedcatalog.NewHTTPValidator(cfg.CatalogServiceURL, cfg.CatalogTimeout)
	}
	if cfg.AuthValidationBackend != "http" {
		return nil, fmt.Errorf("AUTH_VALIDATION_BACKEND=%s is not supported", cfg.AuthValidationBackend)
	}
	playbackService := service.NewWithCatalogValidator(repo, producer, catalogValidator, cfg.EnableCatalogValidation)
	authValidator := auth.NewHTTPValidator(cfg.AuthServiceURL, cfg.AuthTimeout)
	playbackHandler := handler.NewHTTPHandler(playbackService, authValidator)

	mux.HandleFunc("/healthz", healthzHandler)
	mux.HandleFunc("/swagger", swaggerUIHandler)
	mux.HandleFunc("/swagger/", swaggerUIHandler)
	mux.HandleFunc("/swagger/openapi.json", openAPISpecHandler)
	mux.Handle("/v1/jam/sessions/", playbackHandler)
	return mux, nil
}

func swaggerUIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path != "/swagger" && r.URL.Path != "/swagger/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(swaggerUIHTML))
}

func openAPISpecHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(openAPISpec))
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payload := healthResponse{
		Status:    "ok",
		Service:   "playback-service",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}
