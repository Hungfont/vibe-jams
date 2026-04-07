package gateway

import (
	"encoding/json"
	"fmt"
)

const gatewaySwaggerUIHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>API Gateway Swagger</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: '/swagger/openapi.json',
      dom_id: '#swagger-ui',
      deepLinking: true,
      presets: [SwaggerUIBundle.presets.apis],
      layout: 'BaseLayout'
    });
  </script>
</body>
</html>`

func gatewayOpenAPISpec() map[string]any {
	return map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "API Gateway",
			"version":     "1.0.0",
			"description": "Public ingress for auth lifecycle and BFF-first delegated API route families.",
		},
		"servers": []map[string]string{{"url": "http://localhost:8085"}},
		"paths": map[string]any{
			"/healthz": map[string]any{
				"get": map[string]any{
					"tags":        []string{"gateway"},
					"summary":     "Gateway health check",
					"operationId": "getGatewayHealth",
					"responses": map[string]any{
						"200": map[string]any{"description": "Gateway is healthy"},
					},
				},
			},
			"/v1/auth/login":   authRouteSpec("post", "Authenticate user and issue session tokens"),
			"/v1/auth/refresh": authRouteSpec("post", "Refresh session tokens"),
			"/v1/auth/logout":  authRouteSpec("post", "Revoke session"),
			"/v1/auth/me":      authRouteSpec("get", "Return current auth claims"),
			"/v1/bff/mvp/sessions/{sessionId}/orchestration": map[string]any{
				"post": protectedProxySpec(
					"Run BFF orchestration through gateway",
					[]map[string]any{{
						"name":        "sessionId",
						"in":          "path",
						"required":    true,
						"description": "Jam session identifier",
						"schema":      map[string]string{"type": "string"},
					}},
				),
			},
			"/v1/bff/mvp/realtime/ws-config": map[string]any{
				"get": protectedProxySpec(
					"Resolve realtime websocket bootstrap config through gateway",
					[]map[string]any{
						{
							"name":        "sessionId",
							"in":          "query",
							"required":    true,
							"description": "Jam session identifier",
							"schema":      map[string]string{"type": "string"},
						},
						{
							"name":        "lastSeenVersion",
							"in":          "query",
							"required":    false,
							"description": "Client cursor for incremental stream",
							"schema":      map[string]string{"type": "string"},
						},
					},
				),
			},
			"/v1/bff/mvp/realtime/ws": map[string]any{
				"get": protectedProxySpec(
					"Proxy realtime websocket connect path through gateway/BFF",
					[]map[string]any{
						{
							"name":        "sessionId",
							"in":          "query",
							"required":    true,
							"description": "Jam session identifier",
							"schema":      map[string]string{"type": "string"},
						},
						{
							"name":        "lastSeenVersion",
							"in":          "query",
							"required":    false,
							"description": "Client cursor for incremental stream",
							"schema":      map[string]string{"type": "string"},
						},
					},
				),
			},
			"/api/v1/jams/{jamId}/state": map[string]any{
				"get": protectedProxySpec(
					"Delegated jam state route family",
					[]map[string]any{{
						"name":        "jamId",
						"in":          "path",
						"required":    true,
						"description": "Jam identifier",
						"schema":      map[string]string{"type": "string"},
					}},
				),
			},
			"/v1/jam/sessions/{jamId}/playback/commands": map[string]any{
				"post": protectedProxySpec(
					"Delegated playback command route family",
					[]map[string]any{{
						"name":        "jamId",
						"in":          "path",
						"required":    true,
						"description": "Jam identifier",
						"schema":      map[string]string{"type": "string"},
					}},
				),
			},
			"/internal/v1/catalog/tracks/{trackId}": map[string]any{
				"get": protectedProxySpec(
					"Delegated catalog lookup route family",
					[]map[string]any{{
						"name":        "trackId",
						"in":          "path",
						"required":    true,
						"description": "Track identifier",
						"schema":      map[string]string{"type": "string"},
					}},
				),
			},
		},
		"components": map[string]any{
			"securitySchemes": map[string]any{
				"bearerAuth": map[string]string{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "JWT",
				},
			},
		},
	}
}

func authRouteSpec(method, summary string) map[string]any {
	return map[string]any{
		method: map[string]any{
			"tags":        []string{"auth"},
			"summary":     summary,
			"operationId": fmt.Sprintf("%sAuthRoute", method),
			"responses": map[string]any{
				"200": map[string]any{"description": "Auth upstream response"},
				"401": gatewayErrorResponse("Unauthorized"),
			},
		},
	}
}

func protectedProxySpec(summary string, parameters []map[string]any) map[string]any {
	return map[string]any{
		"tags":        []string{"bff-proxy"},
		"summary":     summary,
		"security":    []map[string][]string{{"bearerAuth": []string{}}},
		"parameters":  parameters,
		"operationId": fmt.Sprintf("proxy_%x", summary),
		"responses": map[string]any{
			"200": map[string]any{"description": "Delegated upstream success"},
			"401": gatewayErrorResponse("Missing or invalid credentials"),
			"503": gatewayErrorResponse("Upstream unavailable"),
		},
	}
}

func gatewayErrorResponse(description string) map[string]any {
	return map[string]any{
		"description": description,
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": map[string]string{"type": "object"},
			},
		},
	}
}

func marshalGatewayOpenAPISpec() ([]byte, error) {
	body, err := json.Marshal(gatewayOpenAPISpec())
	if err != nil {
		return nil, fmt.Errorf("marshal gateway openapi: %w", err)
	}
	return body, nil
}
