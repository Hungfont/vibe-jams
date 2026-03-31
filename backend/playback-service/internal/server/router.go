package server

import (
	"encoding/json"
	"net/http"
	"time"

	"video-streaming/backend/playback-service/internal/auth"
	"video-streaming/backend/playback-service/internal/config"
	"video-streaming/backend/playback-service/internal/handler"
	"video-streaming/backend/playback-service/internal/kafka"
	"video-streaming/backend/playback-service/internal/repository"
	"video-streaming/backend/playback-service/internal/service"
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
          "clientEventId": { "type": "string" },
          "expectedQueueVersion": { "type": "integer", "format": "int64" },
          "positionMs": { "type": "integer", "format": "int64" }
        },
        "required": ["command", "clientEventId", "expectedQueueVersion"]
      },
      "CommandAcceptedResponse": {
        "type": "object",
        "properties": {
          "accepted": { "type": "boolean" }
        },
        "required": ["accepted"]
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

// NewRouter builds all playback-service HTTP routes.
func NewRouter(cfg config.Config) http.Handler {
	mux := http.NewServeMux()

	repo := repository.NewRedisPlaybackRepository()
	// Local seed keeps a testable baseline session for manual smoke checks.
	_ = repo.SeedSession("jam-local", "host-local", 1)

	publisher := &kafka.InMemoryPublisher{}
	producer := kafka.NewProducer(publisher)
	playbackService := service.New(repo, producer)
	authValidator := auth.NewHTTPValidator(cfg.AuthServiceURL, cfg.AuthTimeout)
	playbackHandler := handler.NewHTTPHandler(playbackService, authValidator)

	mux.HandleFunc("/healthz", healthzHandler)
	mux.HandleFunc("/swagger", swaggerUIHandler)
	mux.HandleFunc("/swagger/", swaggerUIHandler)
	mux.HandleFunc("/swagger/openapi.json", openAPISpecHandler)
	mux.Handle("/v1/jam/sessions/", playbackHandler)
	return mux
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
