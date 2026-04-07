package bff

import (
	"encoding/json"
	"fmt"
)

const swaggerUIHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>API Service BFF Swagger</title>
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

func openAPISpec() map[string]any {
	paths := map[string]any{
		"/v1/bff/mvp/sessions/{sessionId}/orchestration": map[string]any{
			"post": map[string]any{
				"tags":        []string{"bff"},
				"summary":     "Run MVP orchestration flow",
				"description": "Validates auth, fetches jam state, and optionally enriches with catalog lookup.",
				"operationId": "postSessionOrchestration",
				"security":    []map[string][]string{{"bearerAuth": []string{}}},
				"parameters": []map[string]any{
					{
						"name":        "sessionId",
						"in":          "path",
						"required":    true,
						"description": "Jam session identifier",
						"schema":      map[string]string{"type": "string"},
					},
				},
				"requestBody": map[string]any{
					"required": false,
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]string{"$ref": "#/components/schemas/OrchestrateRequest"},
							"example": map[string]any{
								"trackId": "trk_1",
							},
						},
					},
				},
				"responses": map[string]any{
					"200": map[string]any{
						"description": "Orchestration success (full or partial)",
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]string{"$ref": "#/components/schemas/SuccessEnvelope"},
							},
						},
					},
					"400": errorResponse("Invalid request input"),
					"401": errorResponse("Unauthorized"),
					"404": errorResponse("Required dependency not found"),
					"503": errorResponse("Dependency timeout or unavailable"),
				},
			},
		},
	}

	for route, spec := range delegatedBFFRouteSpecs() {
		paths[route] = spec
	}

	return map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "API Service BFF",
			"version":     "1.0.0",
			"description": "MVP orchestration entrypoint across auth, jam, and optional catalog dependencies.",
		},
		"servers": []map[string]string{{"url": "http://localhost:8084"}},
		"paths":   paths,
		"components": map[string]any{
			"securitySchemes": map[string]any{
				"bearerAuth": map[string]string{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "JWT",
				},
			},
			"schemas": map[string]any{
				"OrchestrateRequest": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"trackId": map[string]string{"type": "string"},
					},
				},
				"SuccessEnvelope": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"success": map[string]string{"type": "boolean"},
						"data":    map[string]string{"$ref": "#/components/schemas/OrchestrateData"},
					},
					"required": []string{"success", "data"},
				},
				"OrchestrateData": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"claims":             map[string]any{"type": "object"},
						"sessionState":       map[string]any{"type": "object"},
						"track":              map[string]any{"type": "object", "nullable": true},
						"partial":            map[string]string{"type": "boolean"},
						"dependencyStatuses": map[string]any{"type": "object", "additionalProperties": map[string]string{"type": "string"}},
						"issues":             map[string]any{"type": "array", "items": map[string]string{"$ref": "#/components/schemas/DependencyIssue"}},
					},
					"required": []string{"claims", "sessionState", "partial", "dependencyStatuses"},
				},
				"DependencyIssue": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"dependency": map[string]string{"type": "string"},
						"code":       map[string]string{"type": "string"},
						"message":    map[string]string{"type": "string"},
					},
					"required": []string{"dependency", "code", "message"},
				},
				"ErrorEnvelope": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"success": map[string]string{"type": "boolean"},
						"error":   map[string]string{"$ref": "#/components/schemas/ErrorBody"},
					},
					"required": []string{"success", "error"},
				},
				"ErrorBody": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"code":       map[string]string{"type": "string"},
						"message":    map[string]string{"type": "string"},
						"dependency": map[string]string{"type": "string"},
					},
					"required": []string{"code", "message"},
				},
			},
		},
	}
}

func delegatedBFFRouteSpecs() map[string]any {
	jamIDPathParam := map[string]any{
		"name":        "jamId",
		"in":          "path",
		"required":    true,
		"description": "Jam identifier",
		"schema":      map[string]string{"type": "string"},
	}
	trackIDPathParam := map[string]any{
		"name":        "trackId",
		"in":          "path",
		"required":    true,
		"description": "Track identifier",
		"schema":      map[string]string{"type": "string"},
	}

	return map[string]any{
		"/api/v1/jams/create": bffDelegatedOperation(httpMethodPost, "Create jam session", nil),
		"/api/v1/jams/{jamId}/join": bffDelegatedOperation(
			httpMethodPost,
			"Join jam session",
			[]map[string]any{jamIDPathParam},
		),
		"/api/v1/jams/{jamId}/leave": bffDelegatedOperation(
			httpMethodPost,
			"Leave jam session",
			[]map[string]any{jamIDPathParam},
		),
		"/api/v1/jams/{jamId}/end": bffDelegatedOperation(
			httpMethodPost,
			"End jam session",
			[]map[string]any{jamIDPathParam},
		),
		"/api/v1/jams/{jamId}/state": bffDelegatedOperation(
			httpMethodGet,
			"Get jam session state",
			[]map[string]any{jamIDPathParam},
		),
		"/api/v1/jams/{jamId}/queue/snapshot": bffDelegatedOperation(
			httpMethodGet,
			"Get queue snapshot",
			[]map[string]any{jamIDPathParam},
		),
		"/api/v1/jams/{jamId}/queue/add": bffDelegatedOperation(
			httpMethodPost,
			"Add queue item",
			[]map[string]any{jamIDPathParam},
		),
		"/api/v1/jams/{jamId}/queue/remove": bffDelegatedOperation(
			httpMethodPost,
			"Remove queue item",
			[]map[string]any{jamIDPathParam},
		),
		"/api/v1/jams/{jamId}/queue/reorder": bffDelegatedOperation(
			httpMethodPost,
			"Reorder queue items",
			[]map[string]any{jamIDPathParam},
		),
		"/api/v1/jams/{jamId}/moderation/mute": bffDelegatedOperation(
			httpMethodPost,
			"Mute participant",
			[]map[string]any{jamIDPathParam},
		),
		"/api/v1/jams/{jamId}/moderation/kick": bffDelegatedOperation(
			httpMethodPost,
			"Kick participant",
			[]map[string]any{jamIDPathParam},
		),
		"/api/v1/jams/{jamId}/permissions": bffDelegatedOperation(
			httpMethodPost,
			"Update guest permissions",
			[]map[string]any{jamIDPathParam},
		),
		"/v1/jam/sessions/{jamId}/playback/commands": bffDelegatedOperation(
			httpMethodPost,
			"Send playback command",
			[]map[string]any{jamIDPathParam},
		),
		"/internal/v1/catalog/tracks/{trackId}": bffDelegatedOperation(
			httpMethodGet,
			"Lookup catalog track",
			[]map[string]any{trackIDPathParam},
		),
		"/v1/bff/mvp/realtime/ws-config": bffDelegatedOperation(
			httpMethodGet,
			"Resolve realtime websocket config",
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
					"description": "Client-side cursor version",
					"schema":      map[string]string{"type": "string"},
				},
			},
		),
		"/v1/bff/mvp/realtime/ws": bffDelegatedOperation(
			httpMethodGet,
			"Proxy realtime websocket connect path to rt-gateway",
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
					"description": "Client-side cursor version",
					"schema":      map[string]string{"type": "string"},
				},
			},
		),
	}
}

const (
	httpMethodGet  = "get"
	httpMethodPost = "post"
)

func bffDelegatedOperation(method, summary string, parameters []map[string]any) map[string]any {
	operation := map[string]any{
		"tags":     []string{"bff-delegated"},
		"summary":  summary,
		"security": []map[string][]string{{"bearerAuth": []string{}}},
		"responses": map[string]any{
			"200": map[string]any{"description": "Delegated upstream response"},
			"400": errorResponse("Invalid request input"),
			"401": errorResponse("Unauthorized"),
			"403": errorResponse("Forbidden"),
			"404": errorResponse("Resource not found"),
			"503": errorResponse("Dependency unavailable"),
		},
	}
	if len(parameters) > 0 {
		operation["parameters"] = parameters
	}

	return map[string]any{method: operation}
}

func errorResponse(description string) map[string]any {
	return map[string]any{
		"description": description,
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": map[string]string{"$ref": "#/components/schemas/ErrorEnvelope"},
			},
		},
	}
}

func marshalOpenAPISpec() ([]byte, error) {
	body, err := jsonMarshal(openAPISpec())
	if err != nil {
		return nil, fmt.Errorf("marshal openapi: %w", err)
	}
	return body, nil
}

var jsonMarshal = func(v any) ([]byte, error) {
	return json.Marshal(v)
}
