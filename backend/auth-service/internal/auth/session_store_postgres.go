package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// PostgresSessionStore persists refresh sessions in postgres.
type PostgresSessionStore struct {
	db    *sql.DB
	table string
}

func NewPostgresSessionStore(db *sql.DB) *PostgresSessionStore {
	return &PostgresSessionStore{db: db, table: "auth_refresh_sessions"}
}

func (s *PostgresSessionStore) Create(ctx context.Context, session RefreshSession) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (
			session_id, family_id, user_id, plan, session_state, scope, token_hash,
			replaced_by_hash, expires_at, issued_at, revoked_at, revoke_reason
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`, s.table)
	_, err := s.db.ExecContext(ctx, query,
		session.SessionID,
		session.FamilyID,
		session.UserID,
		session.Plan,
		session.SessionState,
		strings.Join(cloneScope(session.Scope), " "),
		session.TokenHash,
		session.ReplacedByHash,
		session.ExpiresAt.UTC(),
		session.IssuedAt.UTC(),
		session.RevokedAt,
		session.RevokeReason,
	)
	if err != nil {
		return fmt.Errorf("create refresh session: %w", err)
	}
	return nil
}

func (s *PostgresSessionStore) GetByTokenHash(ctx context.Context, tokenHash string) (RefreshSession, error) {
	query := fmt.Sprintf(`
		SELECT session_id, family_id, user_id, plan, session_state, scope, token_hash,
			replaced_by_hash, expires_at, issued_at, revoked_at, revoke_reason
		FROM %s
		WHERE token_hash=$1
	`, s.table)

	var session RefreshSession
	var scopeRaw string
	var replacedBy sql.NullString
	var revokedAt sql.NullTime
	var revokeReason sql.NullString
	row := s.db.QueryRowContext(ctx, query, tokenHash)
	err := row.Scan(
		&session.SessionID,
		&session.FamilyID,
		&session.UserID,
		&session.Plan,
		&session.SessionState,
		&scopeRaw,
		&session.TokenHash,
		&replacedBy,
		&session.ExpiresAt,
		&session.IssuedAt,
		&revokedAt,
		&revokeReason,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RefreshSession{}, ErrSessionNotFound
		}
		return RefreshSession{}, fmt.Errorf("get refresh session: %w", err)
	}

	session.Scope = splitScope(scopeRaw)
	if replacedBy.Valid {
		session.ReplacedByHash = replacedBy.String
	}
	if revokedAt.Valid {
		timestamp := revokedAt.Time
		session.RevokedAt = &timestamp
	}
	if revokeReason.Valid {
		session.RevokeReason = revokeReason.String
	}
	return session, nil
}

func (s *PostgresSessionStore) Rotate(ctx context.Context, presentedTokenHash string, replacement RefreshSession, now time.Time) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	query := fmt.Sprintf(`
		SELECT replaced_by_hash, revoked_at, expires_at
		FROM %s
		WHERE token_hash=$1
		FOR UPDATE
	`, s.table)
	var replacedBy sql.NullString
	var revokedAt sql.NullTime
	var expiresAt time.Time
	err = tx.QueryRowContext(ctx, query, presentedTokenHash).Scan(&replacedBy, &revokedAt, &expiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrSessionNotFound
		}
		return fmt.Errorf("load refresh session for rotation: %w", err)
	}
	if revokedAt.Valid {
		if replacedBy.Valid && strings.TrimSpace(replacedBy.String) != "" {
			return ErrSessionReused
		}
		return ErrSessionRevoked
	}
	if now.After(expiresAt) {
		return ErrSessionExpired
	}

	updateQuery := fmt.Sprintf(`
		UPDATE %s
		SET revoked_at=$1, revoke_reason=$2, replaced_by_hash=$3
		WHERE token_hash=$4
	`, s.table)
	if _, err = tx.ExecContext(ctx, updateQuery, now.UTC(), "rotated", replacement.TokenHash, presentedTokenHash); err != nil {
		return fmt.Errorf("revoke old refresh session: %w", err)
	}

	insertQuery := fmt.Sprintf(`
		INSERT INTO %s (
			session_id, family_id, user_id, plan, session_state, scope, token_hash,
			replaced_by_hash, expires_at, issued_at, revoked_at, revoke_reason
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`, s.table)
	if _, err = tx.ExecContext(ctx, insertQuery,
		replacement.SessionID,
		replacement.FamilyID,
		replacement.UserID,
		replacement.Plan,
		replacement.SessionState,
		strings.Join(cloneScope(replacement.Scope), " "),
		replacement.TokenHash,
		replacement.ReplacedByHash,
		replacement.ExpiresAt.UTC(),
		replacement.IssuedAt.UTC(),
		replacement.RevokedAt,
		replacement.RevokeReason,
	); err != nil {
		return fmt.Errorf("insert rotated refresh session: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit rotate refresh session: %w", err)
	}
	return nil
}

func (s *PostgresSessionStore) RevokeFamily(ctx context.Context, familyID string, now time.Time, reason string) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET revoked_at=$1, revoke_reason=$2
		WHERE family_id=$3 AND revoked_at IS NULL
	`, s.table)
	_, err := s.db.ExecContext(ctx, query, now.UTC(), reason, familyID)
	if err != nil {
		return fmt.Errorf("revoke refresh family: %w", err)
	}
	return nil
}

func splitScope(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	return strings.Fields(raw)
}
