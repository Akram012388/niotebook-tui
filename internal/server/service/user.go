package service

import (
	"context"
	"unicode/utf8"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/server/store"
)

type UserService struct {
	users store.UserStore
}

func NewUserService(users store.UserStore) *UserService {
	return &UserService{users: users}
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	return s.users.GetUserByID(ctx, id)
}

func (s *UserService) UpdateUser(ctx context.Context, id string, updates *models.UserUpdate) (*models.User, error) {
	if updates.DisplayName != nil {
		if utf8.RuneCountInString(*updates.DisplayName) > 50 || *updates.DisplayName == "" {
			return nil, &models.APIError{
				Code: models.ErrCodeValidation, Field: "display_name",
				Message: "Display name must be 1-50 characters",
			}
		}
	}
	if updates.Bio != nil {
		if utf8.RuneCountInString(*updates.Bio) > 160 {
			return nil, &models.APIError{
				Code: models.ErrCodeValidation, Field: "bio",
				Message: "Bio must be 160 characters or fewer",
			}
		}
	}
	return s.users.UpdateUser(ctx, id, updates)
}
