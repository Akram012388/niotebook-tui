package service_test

import (
	"context"
	"strings"
	"testing"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/server/service"
)

func TestGetUserByID(t *testing.T) {
	userStore := newMockUserStore()
	svc := service.NewUserService(userStore)

	// Seed a user via mock
	userStore.CreateUser(context.Background(), "akram", "akram@example.com", "hash", "Akram")

	user, err := svc.GetUserByID(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if user.Username != "akram" {
		t.Errorf("username = %q, want %q", user.Username, "akram")
	}
}

func TestGetUserByIDNotFound(t *testing.T) {
	userStore := newMockUserStore()
	svc := service.NewUserService(userStore)

	_, err := svc.GetUserByID(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent user")
	}
	apiErr, ok := err.(*models.APIError)
	if !ok {
		t.Fatalf("expected *models.APIError, got %T", err)
	}
	if apiErr.Code != models.ErrCodeNotFound {
		t.Errorf("code = %q, want %q", apiErr.Code, models.ErrCodeNotFound)
	}
}

func TestUpdateUserDisplayName(t *testing.T) {
	userStore := newMockUserStore()
	svc := service.NewUserService(userStore)

	userStore.CreateUser(context.Background(), "akram", "akram@example.com", "hash", "Akram")

	newName := "New Display Name"
	user, err := svc.UpdateUser(context.Background(), "user-1", &models.UserUpdate{
		DisplayName: &newName,
	})
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	if user.DisplayName != "New Display Name" {
		t.Errorf("display_name = %q, want %q", user.DisplayName, "New Display Name")
	}
}

func TestUpdateUserDisplayNameEmpty(t *testing.T) {
	userStore := newMockUserStore()
	svc := service.NewUserService(userStore)

	userStore.CreateUser(context.Background(), "akram", "akram@example.com", "hash", "Akram")

	empty := ""
	_, err := svc.UpdateUser(context.Background(), "user-1", &models.UserUpdate{
		DisplayName: &empty,
	})
	if err == nil {
		t.Fatal("expected validation error for empty display name")
	}
	apiErr, ok := err.(*models.APIError)
	if !ok {
		t.Fatalf("expected *models.APIError, got %T", err)
	}
	if apiErr.Field != "display_name" {
		t.Errorf("field = %q, want %q", apiErr.Field, "display_name")
	}
}

func TestUpdateUserDisplayNameTooLong(t *testing.T) {
	userStore := newMockUserStore()
	svc := service.NewUserService(userStore)

	userStore.CreateUser(context.Background(), "akram", "akram@example.com", "hash", "Akram")

	long := strings.Repeat("a", 51)
	_, err := svc.UpdateUser(context.Background(), "user-1", &models.UserUpdate{
		DisplayName: &long,
	})
	if err == nil {
		t.Fatal("expected validation error for too-long display name")
	}
	apiErr, ok := err.(*models.APIError)
	if !ok {
		t.Fatalf("expected *models.APIError, got %T", err)
	}
	if apiErr.Field != "display_name" {
		t.Errorf("field = %q, want %q", apiErr.Field, "display_name")
	}
}

func TestUpdateUserBioTooLong(t *testing.T) {
	userStore := newMockUserStore()
	svc := service.NewUserService(userStore)

	userStore.CreateUser(context.Background(), "akram", "akram@example.com", "hash", "Akram")

	long := strings.Repeat("b", 161)
	_, err := svc.UpdateUser(context.Background(), "user-1", &models.UserUpdate{
		Bio: &long,
	})
	if err == nil {
		t.Fatal("expected validation error for too-long bio")
	}
	apiErr, ok := err.(*models.APIError)
	if !ok {
		t.Fatalf("expected *models.APIError, got %T", err)
	}
	if apiErr.Field != "bio" {
		t.Errorf("field = %q, want %q", apiErr.Field, "bio")
	}
}

func TestUpdateUserBioValid(t *testing.T) {
	userStore := newMockUserStore()
	svc := service.NewUserService(userStore)

	userStore.CreateUser(context.Background(), "akram", "akram@example.com", "hash", "Akram")

	bio := "This is my bio"
	user, err := svc.UpdateUser(context.Background(), "user-1", &models.UserUpdate{
		Bio: &bio,
	})
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	if user.Bio != "This is my bio" {
		t.Errorf("bio = %q, want %q", user.Bio, "This is my bio")
	}
}

func TestUpdateUserPartialUpdate(t *testing.T) {
	userStore := newMockUserStore()
	svc := service.NewUserService(userStore)

	userStore.CreateUser(context.Background(), "akram", "akram@example.com", "hash", "Akram")

	// Update only bio, display name should remain unchanged
	bio := "New bio"
	user, err := svc.UpdateUser(context.Background(), "user-1", &models.UserUpdate{
		Bio: &bio,
	})
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	if user.DisplayName != "Akram" {
		t.Errorf("display_name = %q, want %q (unchanged)", user.DisplayName, "Akram")
	}
	if user.Bio != "New bio" {
		t.Errorf("bio = %q, want %q", user.Bio, "New bio")
	}
}
