package server

import (
	"encoding/json"
	"net/http"
	"time"

	"video-streaming/backend/jams/internal/auth"
	"video-streaming/backend/jams/internal/config"
	"video-streaming/backend/jams/internal/handler"
	"video-streaming/backend/jams/internal/kafka"
	"video-streaming/backend/jams/internal/repository"
	"video-streaming/backend/jams/internal/service"
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
          "itemId": { "type": "string" }
        },
        "required": ["itemId"]
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
          "message": { "type": "string" }
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
func NewRouter(cfg config.Config) http.Handler {
	mux := http.NewServeMux()
	queueRepo := repository.NewRedisQueueRepository()
	eventPublisher := &kafka.InMemoryPublisher{}
	eventProducer := kafka.NewProducer(eventPublisher)
	queueService := service.New(queueRepo, eventProducer)
	authValidator := auth.NewHTTPValidator(cfg.AuthServiceURL, cfg.AuthTimeout)
	jamsHandler := handler.NewHTTPHandler(queueService, authValidator)
	mux.HandleFunc("/healthz", healthzHandler)
	mux.HandleFunc("/swagger", swaggerUIHandler)
	mux.HandleFunc("/swagger/", swaggerUIHandler)
	mux.HandleFunc("/swagger/openapi.json", openAPISpecHandler)
	mux.Handle("/api/v1/jams/", jamsHandler)
	return mux
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
