package server

import (
  "bufio"
	"encoding/json"
  "errors"
	"fmt"
	"log/slog"
  "net"
	"net/http"
	"strings"
	"time"

	"video-streaming/backend/jams/internal/auth"
	"video-streaming/backend/jams/internal/config"
	"video-streaming/backend/jams/internal/handler"
	"video-streaming/backend/jams/internal/kafka"
	"video-streaming/backend/jams/internal/repository"
	"video-streaming/backend/jams/internal/service"
	sharedauth "video-streaming/backend/shared/auth"
	sharedcatalog "video-streaming/backend/shared/catalog"
	sharedkafka "video-streaming/backend/shared/kafka"
)

const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Video Streaming API Docs</title>
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
    "title": "Video Streaming Backend API",
    "version": "1.0.0",
    "description": "API documentation for local testing."
  },
  "servers": [
    {
      "url": "http://localhost:8080"
    }
  ],
  "paths": {
    "/healthz": {
      "get": {
        "summary": "Health check",
        "responses": {
          "200": {
            "description": "Service is healthy",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "status": { "type": "string", "example": "ok" },
                    "service": { "type": "string", "example": "video-streaming-backend" },
                    "timestamp": { "type": "string", "format": "date-time", "example": "2026-01-01T00:00:00Z" }
                  },
                  "required": ["status", "service", "timestamp"]
                }
              }
            }
          },
          "405": {
            "description": "Method not allowed"
          }
        }
      }
    },
    "/api/v1/jams/create": {
      "post": {
        "summary": "Create a jam session",
        "responses": {
          "201": {
            "description": "Session created",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/SessionSnapshot" }
              }
            }
          },
          "401": { "description": "Unauthorized", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorBody" } } } },
          "403": { "description": "Premium required", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorBody" } } } }
        }
      }
    },
    "/api/v1/jams/{jamId}/join": {
      "post": {
        "summary": "Join an active jam session",
        "parameters": [
          {
            "name": "jamId",
            "in": "path",
            "required": true,
            "schema": { "type": "string" }
          }
        ],
        "responses": {
          "200": {
            "description": "Joined session",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/SessionSnapshot" }
              }
            }
          },
          "401": { "description": "Unauthorized", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorBody" } } } },
          "404": { "description": "Session not found", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorBody" } } } },
          "409": { "description": "Session ended", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorBody" } } } }
        }
      }
    },
    "/api/v1/jams/{jamId}/leave": {
      "post": {
        "summary": "Leave an active jam session (host leave ends session)",
        "parameters": [
          {
            "name": "jamId",
            "in": "path",
            "required": true,
            "schema": { "type": "string" }
          }
        ],
        "responses": {
          "200": {
            "description": "Left session",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/SessionSnapshot" }
              }
            }
          },
          "401": { "description": "Unauthorized", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorBody" } } } },
          "404": { "description": "Session or participant not found", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorBody" } } } },
          "409": { "description": "Session ended", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorBody" } } } }
        }
      }
    },
    "/api/v1/jams/{jamId}/end": {
      "post": {
        "summary": "End an active jam session (host only)",
        "parameters": [
          {
            "name": "jamId",
            "in": "path",
            "required": true,
            "schema": { "type": "string" }
          }
        ],
        "responses": {
          "200": {
            "description": "Ended session",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/SessionSnapshot" }
              }
            }
          },
          "401": { "description": "Unauthorized", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorBody" } } } },
          "403": { "description": "Premium required or host-only constraint", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorBody" } } } },
          "404": { "description": "Session not found", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorBody" } } } },
          "409": { "description": "Session ended", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorBody" } } } }
        }
      }
    },
    "/api/v1/jams/{jamId}/queue/add": {
      "post": {
        "summary": "Add one item to jam queue",
        "parameters": [
          {
            "name": "jamId",
            "in": "path",
            "required": true,
            "schema": { "type": "string" }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/AddQueueItemRequest" }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Queue snapshot after add",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/QueueSnapshot" }
              }
            }
          },
          "400": {
            "description": "Invalid input",
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
    },
    "/api/v1/jams/{jamId}/queue/remove": {
      "post": {
        "summary": "Remove one queue item by itemId",
        "parameters": [
          {
            "name": "jamId",
            "in": "path",
            "required": true,
            "schema": { "type": "string" }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/RemoveQueueItemRequest" }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Queue snapshot after remove",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/QueueSnapshot" }
              }
            }
          },
          "400": {
            "description": "Invalid input",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/ErrorBody" }
              }
            }
          },
          "404": {
            "description": "Queue item not found",
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
    },
    "/api/v1/jams/{jamId}/queue/reorder": {
      "post": {
        "summary": "Reorder queue with optimistic concurrency",
        "parameters": [
          {
            "name": "jamId",
            "in": "path",
            "required": true,
            "schema": { "type": "string" }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/ReorderQueueRequest" }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Queue snapshot after reorder",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/QueueSnapshot" }
              }
            }
          },
          "400": {
            "description": "Invalid input",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/ErrorBody" }
              }
            }
          },
          "404": {
            "description": "Queue item not found",
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
    },
    "/api/v1/jams/{jamId}/queue/snapshot": {
      "get": {
        "summary": "Get latest queue snapshot",
        "parameters": [
          {
            "name": "jamId",
            "in": "path",
            "required": true,
            "schema": { "type": "string" }
          }
        ],
        "responses": {
          "200": {
            "description": "Current queue snapshot",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/QueueSnapshot" }
              }
            }
          },
          "400": {
            "description": "Invalid input",
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
      "QueueItem": {
        "type": "object",
        "properties": {
          "itemId": { "type": "string" },
          "trackId": { "type": "string" },
          "addedBy": { "type": "string" }
        },
        "required": ["itemId", "trackId", "addedBy"]
      },
      "QueueSnapshot": {
        "type": "object",
        "properties": {
          "jamId": { "type": "string" },
          "queueVersion": { "type": "integer", "format": "int64" },
          "items": {
            "type": "array",
            "items": { "$ref": "#/components/schemas/QueueItem" }
          }
        },
        "required": ["jamId", "queueVersion", "items"]
      },
      "SessionParticipant": {
        "type": "object",
        "properties": {
          "userId": { "type": "string" },
          "role": { "type": "string", "example": "host" }
        },
        "required": ["userId", "role"]
      },
      "SessionSnapshot": {
        "type": "object",
        "properties": {
          "jamId": { "type": "string" },
          "status": { "type": "string", "example": "active" },
          "hostUserId": { "type": "string" },
          "sessionVersion": { "type": "integer", "format": "int64" },
          "participants": {
            "type": "array",
            "items": { "$ref": "#/components/schemas/SessionParticipant" }
          },
          "endCause": { "type": "string" },
          "endedBy": { "type": "string" }
        },
        "required": ["jamId", "status", "hostUserId", "sessionVersion", "participants"]
      },
      "AddQueueItemRequest": {
        "type": "object",
        "properties": {
          "trackId": { "type": "string" },
          "addedBy": { "type": "string" },
          "idempotencyKey": { "type": "string" }
        },
        "required": ["trackId", "addedBy", "idempotencyKey"]
      },
      "RemoveQueueItemRequest": {
        "type": "object",
        "properties": {
          "itemId": { "type": "string" },
          "expectedQueueVersion": { "type": "integer", "format": "int64" }
        },
        "required": ["itemId", "expectedQueueVersion"]
      },
      "ReorderQueueRequest": {
        "type": "object",
        "properties": {
          "itemIds": {
            "type": "array",
            "items": { "type": "string" }
          },
          "expectedQueueVersion": { "type": "integer", "format": "int64" }
        },
        "required": ["itemIds", "expectedQueueVersion"]
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

// NewRouter builds all HTTP routes for the backend service.
func NewRouter(cfg config.Config) (http.Handler, error) {
	if err := cfg.ValidateRuntimePolicy(); err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	var queueRepo *repository.RedisQueueRepository
	switch strings.ToLower(cfg.StateStoreBackend) {
	case "inmemory":
		queueRepo = repository.NewRedisQueueRepository()
	case "redis", "postgres":
		durableRepo, err := repository.NewDurableQueueRepository(cfg.StateStorePath)
		if err != nil {
			return nil, err
		}
		queueRepo = durableRepo
	default:
		return nil, fmt.Errorf("unsupported STATE_STORE_BACKEND: %s", cfg.StateStoreBackend)
	}

	var eventPublisher kafka.Publisher
	switch strings.ToLower(cfg.KafkaTransport) {
	case "kafka":
		publisher, err := kafka.NewKafkaPublisher(cfg.KafkaBootstrapServers)
		if err != nil {
			return nil, err
		}
		eventPublisher = publisher
	case "inmemory":
		eventPublisher = &kafka.InMemoryPublisher{}
	default:
		eventPublisher = sharedkafka.NewNoOpsProducer()
		return nil, fmt.Errorf("unsupported KAFKA_TRANSPORT: %s", cfg.KafkaTransport)
	}

	eventProducer := kafka.NewProducer(eventPublisher)
	var catalogValidator sharedcatalog.Validator
	if cfg.EnableCatalogValidation {
		catalogValidator = sharedcatalog.NewHTTPValidator(cfg.CatalogServiceURL, cfg.CatalogTimeout)
	}
	var authValidator auth.Validator
	switch cfg.AuthValidationBackend {
	case "jwt":
		previousKeys, err := sharedauth.ParsePreviousKeys(cfg.JWTPreviousKeys)
		if err != nil {
			return nil, fmt.Errorf("parse JWT previous keys: %w", err)
		}
		verifier, err := sharedauth.NewTokenVerifier(
			sharedauth.VerifierKey{KeyID: cfg.JWTActiveKeyID, Secret: cfg.JWTActiveKeySecret},
			previousKeys,
		)
		if err != nil {
			return nil, fmt.Errorf("build JWT verifier: %w", err)
		}
		authValidator = auth.NewJWTValidator(verifier)
	case "http":
		authValidator = auth.NewHTTPValidator(cfg.AuthServiceURL, cfg.AuthTimeout)
	default:
		return nil, fmt.Errorf("AUTH_VALIDATION_BACKEND=%s is not supported", cfg.AuthValidationBackend)
	}
	queueService := service.NewWithCatalogValidator(queueRepo, eventProducer, catalogValidator, cfg.EnableCatalogValidation)
	jamsHandler := handler.NewHTTPHandler(queueService, authValidator)
	mux.HandleFunc("/healthz", healthzHandler)
	mux.HandleFunc("/swagger", swaggerUIHandler)
	mux.HandleFunc("/swagger/", swaggerUIHandler)
	mux.HandleFunc("/swagger/openapi.json", openAPISpecHandler)
	mux.Handle("/api/v1/jams/", jamsHandler)
	return requestLoggingMiddleware("jam-service", mux), nil
}

// swaggerUIHandler serves Swagger UI for interactive API testing.
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
	if _, err := w.Write([]byte(swaggerUIHTML)); err != nil {
		return
	}
}

// openAPISpecHandler serves the OpenAPI specification for Swagger UI.
func openAPISpecHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(openAPISpec)); err != nil {
		return
	}
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payload := healthResponse{
		Status:    "ok",
		Service:   "video-streaming-backend",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(body); err != nil {
		return
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *loggingResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *loggingResponseWriter) Flush() {
  if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
    flusher.Flush()
  }
}

func (w *loggingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
  hijacker, ok := w.ResponseWriter.(http.Hijacker)
  if !ok {
    return nil, nil, errors.New("response writer does not support hijacking")
  }
  return hijacker.Hijack()
}

func requestLoggingMiddleware(service string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(recorder, r)
		slog.Info("http request",
			"service", service,
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.statusCode,
			"durationMs", time.Since(start).Milliseconds(),
		)
	})
}
