package store_test

import (
	"context"
	"testing"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/server/store"
)

func TestCreateUser(t *testing.T) {
	pool := setupTestDB(t)
	s := store.NewUserStore(pool)
	ctx := context.Background()

	user, err := s.CreateUser(ctx, "testuser", "test@example.com", "$2a$12$fakehash", "testuser")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if user.Username != "testuser" {
		t.Errorf("username = %q, want %q", user.Username, "testuser")
	}
	if user.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestCreateUserDuplicateUsername(t *testing.T) {
	pool := setupTestDB(t)
	s := store.NewUserStore(pool)
	ctx := context.Background()

	_, err := s.CreateUser(ctx, "testuser", "a@example.com", "$2a$12$fakehash", "testuser")
	if err != nil {
		t.Fatalf("first create: %v", err)
	}

	_, err = s.CreateUser(ctx, "testuser", "b@example.com", "$2a$12$fakehash", "testuser")
	if err == nil {
		t.Fatal("expected error for duplicate username")
	}
}

func TestCreateUserDuplicateEmail(t *testing.T) {
	pool := setupTestDB(t)
	s := store.NewUserStore(pool)
	ctx := context.Background()

	_, err := s.CreateUser(ctx, "user1", "same@example.com", "$2a$12$fakehash", "user1")
	if err != nil {
		t.Fatalf("first create: %v", err)
	}

	_, err = s.CreateUser(ctx, "user2", "same@example.com", "$2a$12$fakehash", "user2")
	if err == nil {
		t.Fatal("expected error for duplicate email")
	}
}

func TestGetUserByEmail(t *testing.T) {
	pool := setupTestDB(t)
	s := store.NewUserStore(pool)
	ctx := context.Background()

	_, err := s.CreateUser(ctx, "akram", "akram@example.com", "$2a$12$hashvalue", "akram")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	user, hash, err := s.GetUserByEmail(ctx, "akram@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if user.Username != "akram" {
		t.Errorf("username = %q, want %q", user.Username, "akram")
	}
	if hash != "$2a$12$hashvalue" {
		t.Errorf("hash = %q, want %q", hash, "$2a$12$hashvalue")
	}
}

func TestGetUserByEmailNotFound(t *testing.T) {
	pool := setupTestDB(t)
	s := store.NewUserStore(pool)
	ctx := context.Background()

	_, _, err := s.GetUserByEmail(ctx, "nobody@example.com")
	if err == nil {
		t.Fatal("expected error for nonexistent email")
	}
}

func TestGetUserByID(t *testing.T) {
	pool := setupTestDB(t)
	s := store.NewUserStore(pool)
	ctx := context.Background()

	created, _ := s.CreateUser(ctx, "akram", "akram@example.com", "$2a$12$hash", "akram")

	user, err := s.GetUserByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if user.Username != "akram" {
		t.Errorf("username = %q, want %q", user.Username, "akram")
	}
}

func TestGetUserByUsername(t *testing.T) {
	pool := setupTestDB(t)
	s := store.NewUserStore(pool)
	ctx := context.Background()

	_, err := s.CreateUser(ctx, "testuser", "test@example.com", "$2a$12$fakehash", "testuser")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	user, err := s.GetUserByUsername(ctx, "testuser")
	if err != nil {
		t.Fatalf("GetUserByUsername: %v", err)
	}
	if user.Username != "testuser" {
		t.Errorf("username = %q, want %q", user.Username, "testuser")
	}
}

func TestGetUserByUsernameNotFound(t *testing.T) {
	pool := setupTestDB(t)
	s := store.NewUserStore(pool)
	ctx := context.Background()

	_, err := s.GetUserByUsername(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent username")
	}
}

func TestGetUserByIDNotFound(t *testing.T) {
	pool := setupTestDB(t)
	s := store.NewUserStore(pool)
	ctx := context.Background()

	_, err := s.GetUserByID(ctx, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatal("expected error for nonexistent user ID")
	}
}

func TestUpdateUserNoChanges(t *testing.T) {
	pool := setupTestDB(t)
	s := store.NewUserStore(pool)
	ctx := context.Background()

	created, _ := s.CreateUser(ctx, "akram", "akram@example.com", "$2a$12$hash", "akram")

	user, err := s.UpdateUser(ctx, created.ID, &models.UserUpdate{})
	if err != nil {
		t.Fatalf("UpdateUser with no changes: %v", err)
	}
	if user.Username != "akram" {
		t.Errorf("username = %q, want %q", user.Username, "akram")
	}
}

func TestUpdateUserBioOnly(t *testing.T) {
	pool := setupTestDB(t)
	s := store.NewUserStore(pool)
	ctx := context.Background()

	created, _ := s.CreateUser(ctx, "akram", "akram@example.com", "$2a$12$hash", "akram")

	newBio := "Hello world"
	user, err := s.UpdateUser(ctx, created.ID, &models.UserUpdate{Bio: &newBio})
	if err != nil {
		t.Fatalf("UpdateUser bio only: %v", err)
	}
	if user.Bio != "Hello world" {
		t.Errorf("bio = %q, want %q", user.Bio, "Hello world")
	}
}

func TestUpdateUser(t *testing.T) {
	pool := setupTestDB(t)
	s := store.NewUserStore(pool)
	ctx := context.Background()

	created, _ := s.CreateUser(ctx, "akram", "akram@example.com", "$2a$12$hash", "akram")

	newName := "Shaikh Akram"
	newBio := "Building Niotebook."
	updated, err := s.UpdateUser(ctx, created.ID, &models.UserUpdate{
		DisplayName: &newName,
		Bio:         &newBio,
	})
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	if updated.DisplayName != "Shaikh Akram" {
		t.Errorf("display_name = %q, want %q", updated.DisplayName, "Shaikh Akram")
	}
	if updated.Bio != "Building Niotebook." {
		t.Errorf("bio = %q, want %q", updated.Bio, "Building Niotebook.")
	}
}
