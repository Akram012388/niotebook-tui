package store

import (
	"context"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/models"
)

type UserStore interface {
	CreateUser(ctx context.Context, username, email, passwordHash, displayName string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, string, error) // returns user + password hash
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	UpdateUser(ctx context.Context, id string, updates *models.UserUpdate) (*models.User, error)
}

type PostStore interface {
	CreatePost(ctx context.Context, authorID, content string) (*models.Post, error)
	GetPostByID(ctx context.Context, id string) (*models.Post, error)
	GetTimeline(ctx context.Context, cursor time.Time, limit int) ([]models.Post, error)
	GetUserPosts(ctx context.Context, userID string, cursor time.Time, limit int) ([]models.Post, error)
}

type RefreshTokenStore interface {
	StoreToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error
	GetByHash(ctx context.Context, tokenHash string) (id, userID string, expiresAt time.Time, err error)
	DeleteByHash(ctx context.Context, tokenHash string) error
	DeleteAllForUser(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) (int64, error)
}
