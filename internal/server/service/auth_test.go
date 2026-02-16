package service_test

import (
	"context"
	"testing"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/server/service"
)

func TestRegister(t *testing.T) {
	userStore := newMockUserStore()
	tokenStore := newMockRefreshTokenStore()
	auth := service.NewAuthService(userStore, tokenStore, "test-secret-32-bytes-long-xxxxx")

	resp, err := auth.Register(context.Background(), &models.RegisterRequest{
		Username: "akram",
		Email:    "akram@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if resp.User.Username != "akram" {
		t.Errorf("username = %q, want %q", resp.User.Username, "akram")
	}
	if resp.Tokens.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if resp.Tokens.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
}

func TestRegisterInvalidUsername(t *testing.T) {
	userStore := newMockUserStore()
	tokenStore := newMockRefreshTokenStore()
	auth := service.NewAuthService(userStore, tokenStore, "test-secret-32-bytes-long-xxxxx")

	_, err := auth.Register(context.Background(), &models.RegisterRequest{
		Username: "a",
		Email:    "a@example.com",
		Password: "password123",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestRegisterShortPassword(t *testing.T) {
	userStore := newMockUserStore()
	tokenStore := newMockRefreshTokenStore()
	auth := service.NewAuthService(userStore, tokenStore, "test-secret-32-bytes-long-xxxxx")

	_, err := auth.Register(context.Background(), &models.RegisterRequest{
		Username: "akram",
		Email:    "akram@example.com",
		Password: "short",
	})
	if err == nil {
		t.Fatal("expected validation error for short password")
	}
}

func TestLogin(t *testing.T) {
	userStore := newMockUserStore()
	tokenStore := newMockRefreshTokenStore()
	auth := service.NewAuthService(userStore, tokenStore, "test-secret-32-bytes-long-xxxxx")

	// Register first
	if _, err := auth.Register(context.Background(), &models.RegisterRequest{
		Username: "akram", Email: "akram@example.com", Password: "password123",
	}); err != nil {
		t.Fatalf("setup Register: %v", err)
	}

	// Login
	resp, err := auth.Login(context.Background(), &models.LoginRequest{
		Email: "akram@example.com", Password: "password123",
	})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if resp.User.Username != "akram" {
		t.Errorf("username = %q, want %q", resp.User.Username, "akram")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	userStore := newMockUserStore()
	tokenStore := newMockRefreshTokenStore()
	auth := service.NewAuthService(userStore, tokenStore, "test-secret-32-bytes-long-xxxxx")

	if _, err := auth.Register(context.Background(), &models.RegisterRequest{
		Username: "akram", Email: "akram@example.com", Password: "password123",
	}); err != nil {
		t.Fatalf("setup Register: %v", err)
	}

	_, err := auth.Login(context.Background(), &models.LoginRequest{
		Email: "akram@example.com", Password: "wrongpassword",
	})
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	userStore := newMockUserStore()
	tokenStore := newMockRefreshTokenStore()
	auth := service.NewAuthService(userStore, tokenStore, "test-secret-32-bytes-long-xxxxx")

	// Register first user
	if _, err := auth.Register(context.Background(), &models.RegisterRequest{
		Username: "akram", Email: "akram@example.com", Password: "password123",
	}); err != nil {
		t.Fatalf("first Register: %v", err)
	}

	// Register second user with same email
	_, err := auth.Register(context.Background(), &models.RegisterRequest{
		Username: "other", Email: "akram@example.com", Password: "password456",
	})
	if err == nil {
		t.Fatal("expected error for duplicate email")
	}
	apiErr, ok := err.(*models.APIError)
	if !ok {
		t.Fatalf("expected *models.APIError, got %T", err)
	}
	if apiErr.Code != models.ErrCodeConflict {
		t.Errorf("code = %q, want %q", apiErr.Code, models.ErrCodeConflict)
	}
}

func TestLoginNonexistentEmail(t *testing.T) {
	userStore := newMockUserStore()
	tokenStore := newMockRefreshTokenStore()
	auth := service.NewAuthService(userStore, tokenStore, "test-secret-32-bytes-long-xxxxx")

	_, err := auth.Login(context.Background(), &models.LoginRequest{
		Email: "nonexistent@example.com", Password: "password123",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent email")
	}
}

func TestRefreshToken(t *testing.T) {
	userStore := newMockUserStore()
	tokenStore := newMockRefreshTokenStore()
	auth := service.NewAuthService(userStore, tokenStore, "test-secret-32-bytes-long-xxxxx")

	resp, _ := auth.Register(context.Background(), &models.RegisterRequest{
		Username: "akram", Email: "akram@example.com", Password: "password123",
	})

	newTokens, err := auth.Refresh(context.Background(), resp.Tokens.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if newTokens.AccessToken == "" {
		t.Error("expected new access token")
	}
	// Old refresh token should be consumed (single-use)
	_, err = auth.Refresh(context.Background(), resp.Tokens.RefreshToken)
	if err == nil {
		t.Fatal("expected error for reused refresh token")
	}
}
