package service

import (
	"context"

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
		if err := ValidateDisplayName(*updates.DisplayName); err != nil {
			return nil, err
		}
	}
	if updates.Bio != nil {
		if err := ValidateBio(*updates.Bio); err != nil {
			return nil, err
		}
	}
	return s.users.UpdateUser(ctx, id, updates)
}
