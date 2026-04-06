package auth

import (
	"context"
	"strings"
)

// CredentialStore validates user credentials.
type CredentialStore interface {
	Authenticate(ctx context.Context, identity string, password string) (UserRecord, error)
}

// InMemoryCredentialStore provides deterministic fixture users for local and tests.
type InMemoryCredentialStore struct {
	users map[string]inMemoryUser
}

type inMemoryUser struct {
	Password string
	User     UserRecord
}

func NewInMemoryCredentialStore() *InMemoryCredentialStore {
	return &InMemoryCredentialStore{
		users: map[string]inMemoryUser{
			"premium@example.com": {
				Password: "premium-pass",
				User: UserRecord{
					UserID: "user-premium-1",
					Plan:   "premium",
					Scope:  []string{"jam:read", "jam:control", "playback:write"},
				},
			},
			"free@example.com": {
				Password: "free-pass",
				User: UserRecord{
					UserID: "user-free-1",
					Plan:   "free",
					Scope:  []string{"jam:read", "playback:read"},
				},
			},
		},
	}
}

func (s *InMemoryCredentialStore) Authenticate(_ context.Context, identity string, password string) (UserRecord, error) {
	record, ok := s.users[normalizeIdentity(identity)]
	if !ok || record.Password != password {
		return UserRecord{}, ErrInvalidCredentials
	}
	return UserRecord{
		UserID: record.User.UserID,
		Plan:   record.User.Plan,
		Scope:  cloneScope(record.User.Scope),
	}, nil
}

func normalizeIdentity(identity string) string {
	return strings.ToLower(strings.TrimSpace(identity))
}
