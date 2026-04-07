# go-service-swagger-openapi Specification

## Purpose
TBD - created by archiving change gateway-swagger-authn-compat-fix. Update Purpose after archive.
## Requirements
### Requirement: Go services exposing public operational APIs MUST provide Swagger/OpenAPI endpoint documentation
Go services that are public ingress or frontend-facing coordination services SHALL expose a browsable Swagger UI endpoint and an OpenAPI JSON document that describe health, primary route families, and error envelope behaviors.

#### Scenario: API gateway swagger availability
- **WHEN** client requests `GET /swagger` or `GET /swagger/openapi.json` on api-gateway
- **THEN** gateway serves Swagger UI and OpenAPI content describing gateway health and proxy ingress routes

#### Scenario: API-service BFF swagger includes expanded route families
- **WHEN** client requests api-service OpenAPI spec
- **THEN** spec includes orchestration and non-orchestration BFF route families used by frontend BFF-first flow

