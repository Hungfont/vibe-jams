package authz

import (
	"context"
	"strings"

	"video-streaming/backend/jams/internal/model"
	sharedauth "video-streaming/backend/shared/auth"
)

const (
	ActorRoleHost   = "host"
	ActorRoleMember = "member"
	ActorRoleGuest  = "guest"
)

// PolicyCommand identifies permission and moderation command families.
type PolicyCommand string

const (
	CommandModerationMute     PolicyCommand = "moderation.mute"
	CommandModerationKick     PolicyCommand = "moderation.kick"
	CommandPermissionPlayback PolicyCommand = "permission.control_playback"
	CommandPermissionReorder  PolicyCommand = "permission.reorder_queue"
	CommandPermissionVolume   PolicyCommand = "permission.change_volume"
)

// DecisionContext contains auth claim and command context for one policy decision.
type DecisionContext struct {
	JamID        string
	Claims       sharedauth.Claims
	Command      PolicyCommand
	TargetUserID string
}

// Decision is the normalized authorization result returned by guard evaluation.
type Decision struct {
	Allowed   bool
	Reason    string
	ActorRole string
}

// Guard defines centralized policy authorization behavior for jam commands.
type Guard interface {
	Authorize(ctx context.Context, session model.SessionSnapshot, decisionCtx DecisionContext) Decision
}

// HostGuestGuard enforces host-only command authorization semantics.
type HostGuestGuard struct{}

// NewHostGuestGuard creates default host-guest authorization guard.
func NewHostGuestGuard() *HostGuestGuard {
	return &HostGuestGuard{}
}

// Authorize evaluates authorization for one policy command.
func (g *HostGuestGuard) Authorize(_ context.Context, session model.SessionSnapshot, decisionCtx DecisionContext) Decision {
	if err := sharedauth.ValidateClaims(decisionCtx.Claims); err != nil {
		return Decision{Allowed: false, Reason: "invalid_claims", ActorRole: ActorRoleGuest}
	}
	if strings.ToLower(strings.TrimSpace(decisionCtx.Claims.SessionState)) != sharedauth.SessionStateValid {
		return Decision{Allowed: false, Reason: "invalid_session_state", ActorRole: ActorRoleGuest}
	}

	actorRole := resolveActorRole(session, decisionCtx.Claims.UserID)
	if isHostOnlyCommand(decisionCtx.Command) && actorRole != ActorRoleHost {
		return Decision{Allowed: false, Reason: "host_only", ActorRole: actorRole}
	}

	return Decision{Allowed: true, Reason: "allowed", ActorRole: actorRole}
}

func isHostOnlyCommand(command PolicyCommand) bool {
	switch command {
	case CommandModerationMute,
		CommandModerationKick,
		CommandPermissionPlayback,
		CommandPermissionReorder,
		CommandPermissionVolume:
		return true
	default:
		return false
	}
}

func resolveActorRole(session model.SessionSnapshot, actorUserID string) string {
	if strings.TrimSpace(actorUserID) == "" {
		return ActorRoleGuest
	}
	if actorUserID == session.HostUserID {
		return ActorRoleHost
	}
	for _, participant := range session.Participants {
		if participant.UserID == actorUserID {
			if strings.ToLower(string(participant.Role)) == ActorRoleHost {
				return ActorRoleHost
			}
			return ActorRoleMember
		}
	}

	return ActorRoleGuest
}
