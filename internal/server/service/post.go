package service

import (
	"context"
	"strings"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/server/store"
)

type PostService struct {
	posts store.PostStore
}

func NewPostService(posts store.PostStore) *PostService {
	return &PostService{posts: posts}
}

func (s *PostService) CreatePost(ctx context.Context, authorID, content string) (*models.Post, error) {
	content = strings.TrimSpace(content)
	if err := ValidatePostContent(content); err != nil {
		return nil, err
	}
	return s.posts.CreatePost(ctx, authorID, content)
}

func (s *PostService) GetPostByID(ctx context.Context, id string) (*models.Post, error) {
	return s.posts.GetPostByID(ctx, id)
}

func (s *PostService) GetTimeline(ctx context.Context, cursor time.Time, limit int) ([]models.Post, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.posts.GetTimeline(ctx, cursor, limit)
}

func (s *PostService) GetUserPosts(ctx context.Context, userID string, cursor time.Time, limit int) ([]models.Post, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.posts.GetUserPosts(ctx, userID, cursor, limit)
}
