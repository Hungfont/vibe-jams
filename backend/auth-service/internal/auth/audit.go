package auth

import (
	"context"
	"log/slog"
	"time"
)

const (
	AuditEventLoginSuccess      = "login_success"
	AuditEventLoginFailure      = "login_failure"
	AuditEventRefreshSuccess    = "refresh_success"
	AuditEventRefreshFailure    = "refresh_failure"
	AuditEventRefreshReuse      = "refresh_reuse_detected"
	AuditEventLogout            = "logout"
	AuditEventMeUnauthorized    = "me_unauthorized"
	AuditEventValidateFailure   = "validate_failure"
	AuditEventValidateSucceeded = "validate_success"
)

// AuditEvent represents one structured auth lifecycle event.
type AuditEvent struct {
	Event      string
	UserID     string
	Identity   string
	IP         string
	SessionID  string
	FamilyID   string
	Outcome    string
	Reason     string
	OccurredAt time.Time
}

// AuditLogger allows plugging in structured event sinks.
type AuditLogger interface {
	Log(ctx context.Context, event AuditEvent)
}

// SlogAuditLogger writes audit events as structured logs.
type SlogAuditLogger struct {
	logger *slog.Logger
}

func NewSlogAuditLogger(logger *slog.Logger) *SlogAuditLogger {
	if logger == nil {
		logger = slog.Default()
	}
	return &SlogAuditLogger{logger: logger}
}

func (l *SlogAuditLogger) Log(_ context.Context, event AuditEvent) {
	l.logger.Info("auth_audit",
		"event", event.Event,
		"userId", event.UserID,
		"identity", event.Identity,
		"ip", event.IP,
		"sessionId", event.SessionID,
		"familyId", event.FamilyID,
		"outcome", event.Outcome,
		"reason", event.Reason,
		"occurredAt", event.OccurredAt.UTC().Format(time.RFC3339Nano),
	)
}

// NoopAuditLogger is the default audit sink for tests without assertions.
type NoopAuditLogger struct{}

func (NoopAuditLogger) Log(context.Context, AuditEvent) {}
