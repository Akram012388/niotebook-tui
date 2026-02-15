package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type refreshTokenStore struct {
	pool *pgxpool.Pool
}

func NewRefreshTokenStore(pool *pgxpool.Pool) RefreshTokenStore {
	return &refreshTokenStore{pool: pool}
}

func (s *refreshTokenStore) StoreToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("store refresh token: %w", err)
	}
	return nil
}

func (s *refreshTokenStore) GetByHash(ctx context.Context, tokenHash string) (id, userID string, expiresAt time.Time, err error) {
	err = s.pool.QueryRow(ctx,
		`SELECT id, user_id, expires_at
		 FROM refresh_tokens
		 WHERE token_hash = $1`, tokenHash,
	).Scan(&id, &userID, &expiresAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", time.Time{}, &models.APIError{Code: models.ErrCodeUnauthorized, Message: "Invalid refresh token"}
		}
		return "", "", time.Time{}, fmt.Errorf("get refresh token by hash: %w", err)
	}

	return id, userID, expiresAt, nil
}

func (s *refreshTokenStore) DeleteByHash(ctx context.Context, tokenHash string) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM refresh_tokens WHERE token_hash = $1`, tokenHash,
	)
	if err != nil {
		return fmt.Errorf("delete refresh token by hash: %w", err)
	}
	return nil
}

func (s *refreshTokenStore) DeleteAllForUser(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM refresh_tokens WHERE user_id = $1`, userID,
	)
	if err != nil {
		return fmt.Errorf("delete all refresh tokens for user: %w", err)
	}
	return nil
}

func (s *refreshTokenStore) DeleteExpired(ctx context.Context) (int64, error) {
	tag, err := s.pool.Exec(ctx,
		`DELETE FROM refresh_tokens WHERE expires_at < NOW()`,
	)
	if err != nil {
		return 0, fmt.Errorf("delete expired refresh tokens: %w", err)
	}
	return tag.RowsAffected(), nil
}
