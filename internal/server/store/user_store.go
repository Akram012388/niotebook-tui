package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userStore struct {
	pool *pgxpool.Pool
}

func NewUserStore(pool *pgxpool.Pool) UserStore {
	return &userStore{pool: pool}
}

func (s *userStore) CreateUser(ctx context.Context, username, email, passwordHash, displayName string) (*models.User, error) {
	var user models.User
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users (username, email, password, display_name)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, username, display_name, bio, created_at`,
		strings.ToLower(username), strings.ToLower(email), passwordHash, displayName,
	).Scan(&user.ID, &user.Username, &user.DisplayName, &user.Bio, &user.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "username") {
				return nil, &models.APIError{Code: models.ErrCodeConflict, Message: "username already taken", Field: "username"}
			}
			if strings.Contains(pgErr.ConstraintName, "email") {
				return nil, &models.APIError{Code: models.ErrCodeConflict, Message: "email already registered", Field: "email"}
			}
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &user, nil
}

func (s *userStore) GetUserByEmail(ctx context.Context, email string) (*models.User, string, error) {
	var user models.User
	var passwordHash string
	var ignoredEmail string
	err := s.pool.QueryRow(ctx,
		`SELECT id, username, email, password, display_name, bio, created_at
		 FROM users WHERE LOWER(email) = LOWER($1)`, email,
	).Scan(&user.ID, &user.Username, &ignoredEmail, &passwordHash, &user.DisplayName, &user.Bio, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", &models.APIError{Code: models.ErrCodeUnauthorized, Message: "invalid email or password"}
		}
		return nil, "", fmt.Errorf("get user by email: %w", err)
	}

	return &user, passwordHash, nil
}

func (s *userStore) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	err := s.pool.QueryRow(ctx,
		`SELECT id, username, display_name, bio, created_at
		 FROM users WHERE id = $1`, id,
	).Scan(&user.ID, &user.Username, &user.DisplayName, &user.Bio, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &models.APIError{Code: models.ErrCodeNotFound, Message: "user not found"}
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return &user, nil
}

func (s *userStore) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := s.pool.QueryRow(ctx,
		`SELECT id, username, display_name, bio, created_at
		 FROM users WHERE LOWER(username) = LOWER($1)`, username,
	).Scan(&user.ID, &user.Username, &user.DisplayName, &user.Bio, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &models.APIError{Code: models.ErrCodeNotFound, Message: "user not found"}
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}

	return &user, nil
}

func (s *userStore) UpdateUser(ctx context.Context, id string, updates *models.UserUpdate) (*models.User, error) {
	setClauses := []string{}
	args := []any{}
	argIdx := 1

	if updates.DisplayName != nil {
		setClauses = append(setClauses, fmt.Sprintf("display_name = $%d", argIdx))
		args = append(args, *updates.DisplayName)
		argIdx++
	}
	if updates.Bio != nil {
		setClauses = append(setClauses, fmt.Sprintf("bio = $%d", argIdx))
		args = append(args, *updates.Bio)
		argIdx++
	}

	if len(setClauses) == 0 {
		return s.GetUserByID(ctx, id)
	}

	setClauses = append(setClauses, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf(
		`UPDATE users SET %s WHERE id = $%d
		 RETURNING id, username, display_name, bio, created_at`,
		strings.Join(setClauses, ", "), argIdx,
	)

	var user models.User
	err := s.pool.QueryRow(ctx, query, args...).
		Scan(&user.ID, &user.Username, &user.DisplayName, &user.Bio, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	return &user, nil
}
