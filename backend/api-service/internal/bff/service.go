package bff

import (
	"context"
	"errors"
	"log/slog"
	"strings"
)

// Service orchestrates MVP BFF flows.
type Service struct {
	auth           AuthClient
	jam            JamClient
	playback       PlaybackClient
	catalog        CatalogClient
	featureEnabled bool
}

// NewService builds orchestration service.
func NewService(auth AuthClient, jam JamClient, playback PlaybackClient, catalog CatalogClient, featureEnabled bool) *Service {
	return &Service{auth: auth, jam: jam, playback: playback, catalog: catalog, featureEnabled: featureEnabled}
}

// Orchestrate executes read aggregation across auth, jam, and optional catalog for one jam session.
func (s *Service) Orchestrate(ctx context.Context, jamID string, authHeader string, req OrchestrateRequest) (OrchestrateData, *ErrorBody, int) {
	if !s.featureEnabled {
		return OrchestrateData{}, &ErrorBody{Code: "service_unavailable", Message: "bff orchestration is disabled"}, 503
	}
	if strings.TrimSpace(jamID) == "" {
		return OrchestrateData{}, &ErrorBody{Code: "invalid_input", Message: "sessionId is required"}, 400
	}
	if strings.TrimSpace(authHeader) == "" {
		return OrchestrateData{}, &ErrorBody{Code: "unauthorized", Message: "missing authorization"}, 401
	}
	if req.PlaybackCommand != nil {
		return OrchestrateData{}, &ErrorBody{Code: "invalid_input", Message: "playbackCommand is not supported on orchestration endpoint"}, 400
	}

	claims, err := s.auth.ValidateBearerToken(ctx, authHeader)
	if err != nil {
		errBody, status := s.requiredDependencyError("auth", err)
		return OrchestrateData{}, errBody, status
	}

	state, err := s.jam.SessionState(ctx, jamID, authHeader)
	if err != nil {
		errBody, status := s.requiredDependencyError("jam", err)
		return OrchestrateData{}, errBody, status
	}

	data := OrchestrateData{
		Claims:       claims,
		SessionState: state,
		DependencyStatuses: map[string]string{
			"auth": "ok",
			"jam":  "ok",
		},
	}

	if trackID := strings.TrimSpace(req.TrackID); trackID != "" {
		lookup, lookupErr := s.catalog.LookupTrack(ctx, trackID)
		if lookupErr != nil {
			issue := optionalIssue("catalog", lookupErr)
			data.Partial = true
			data.Issues = append(data.Issues, issue)
			data.DependencyStatuses["catalog"] = "degraded"
			slog.Warn("optional dependency failed", "dependency", "catalog", "code", issue.Code, "message", issue.Message)
		} else {
			data.Track = &lookup
			data.DependencyStatuses["catalog"] = "ok"
		}
	}

	return data, nil, 200
}

func (s *Service) requiredDependencyError(dependency string, err error) (*ErrorBody, int) {
	code, message := mapDependencyError(err)
	status := 502
	switch {
	case code == "unauthorized":
		status = 401
	case code == "not_found":
		status = 404
	case code == "dependency_timeout" || code == "dependency_unavailable":
		status = 503
	default:
		var upstream UpstreamError
		if errors.As(err, &upstream) {
			if upstream.StatusCode >= 400 && upstream.StatusCode < 500 {
				status = 424
			} else if upstream.StatusCode >= 500 {
				status = 503
			}
			if strings.TrimSpace(upstream.Code) != "" {
				code = upstream.Code
			}
			if strings.TrimSpace(upstream.Message) != "" {
				message = upstream.Message
			}
		}
	}

	slog.Error("required dependency failed", "dependency", dependency, "code", code, "status", status, "error", err)
	return &ErrorBody{Code: code, Message: message, Dependency: dependency}, status
}

func optionalIssue(dependency string, err error) DependencyIssue {
	code, message := mapDependencyError(err)
	var upstream UpstreamError
	if errors.As(err, &upstream) {
		if strings.TrimSpace(upstream.Code) != "" {
			code = upstream.Code
		}
		if strings.TrimSpace(upstream.Message) != "" {
			message = upstream.Message
		}
	}
	return DependencyIssue{Dependency: dependency, Code: code, Message: message}
}
