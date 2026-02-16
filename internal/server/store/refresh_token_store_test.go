package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/server/store"
)

func TestStoreAndGetToken(t *testing.T) {
	pool := setupTestDB(t)
	us := store.NewUserStore(pool)
	rs := store.NewRefreshTokenStore(pool)
	ctx := context.Background()

	userID := createTestUser(t, us, "akram", "akram@example.com")

	expiresAt := time.Now().Add(24 * time.Hour)
	err := rs.StoreToken(ctx, userID, "abc123hash", expiresAt)
	if err != nil {
		t.Fatalf("StoreToken: %v", err)
	}

	id, gotUserID, gotExpires, err := rs.GetByHash(ctx, "abc123hash")
	if err != nil {
		t.Fatalf("GetByHash: %v", err)
	}
	if id == "" {
		t.Error("expected non-empty token ID")
	}
	if gotUserID != userID {
		t.Errorf("user_id = %q, want %q", gotUserID, userID)
	}
	if gotExpires.Before(time.Now()) {
		t.Error("expected expires_at in the future")
	}
}

func TestDeleteByHash(t *testing.T) {
	pool := setupTestDB(t)
	us := store.NewUserStore(pool)
	rs := store.NewRefreshTokenStore(pool)
	ctx := context.Background()

	userID := createTestUser(t, us, "akram", "akram@example.com")

	_ = rs.StoreToken(ctx, userID, "tokenhash1", time.Now().Add(24*time.Hour))

	err := rs.DeleteByHash(ctx, "tokenhash1")
	if err != nil {
		t.Fatalf("DeleteByHash: %v", err)
	}

	_, _, _, err = rs.GetByHash(ctx, "tokenhash1")
	if err == nil {
		t.Fatal("expected error after deleting token")
	}
}

func TestDeleteAllForUser(t *testing.T) {
	pool := setupTestDB(t)
	us := store.NewUserStore(pool)
	rs := store.NewRefreshTokenStore(pool)
	ctx := context.Background()

	userID := createTestUser(t, us, "akram", "akram@example.com")

	_ = rs.StoreToken(ctx, userID, "hash1", time.Now().Add(24*time.Hour))
	_ = rs.StoreToken(ctx, userID, "hash2", time.Now().Add(24*time.Hour))
	_ = rs.StoreToken(ctx, userID, "hash3", time.Now().Add(24*time.Hour))

	err := rs.DeleteAllForUser(ctx, userID)
	if err != nil {
		t.Fatalf("DeleteAllForUser: %v", err)
	}

	_, _, _, err = rs.GetByHash(ctx, "hash1")
	if err == nil {
		t.Fatal("expected error for deleted token hash1")
	}
	_, _, _, err = rs.GetByHash(ctx, "hash2")
	if err == nil {
		t.Fatal("expected error for deleted token hash2")
	}
}

func TestGetByHashNotFound(t *testing.T) {
	pool := setupTestDB(t)
	s := store.NewRefreshTokenStore(pool)
	ctx := context.Background()

	_, _, _, err := s.GetByHash(ctx, "nonexistent-hash")
	if err == nil {
		t.Fatal("expected error for nonexistent token hash")
	}
}

func TestDeleteExpiredNoTokens(t *testing.T) {
	pool := setupTestDB(t)
	s := store.NewRefreshTokenStore(pool)
	ctx := context.Background()

	deleted, err := s.DeleteExpired(ctx)
	if err != nil {
		t.Fatalf("DeleteExpired: %v", err)
	}
	if deleted != 0 {
		t.Errorf("expected 0 deleted, got %d", deleted)
	}
}

func TestDeleteExpired(t *testing.T) {
	pool := setupTestDB(t)
	us := store.NewUserStore(pool)
	rs := store.NewRefreshTokenStore(pool)
	ctx := context.Background()

	userID := createTestUser(t, us, "akram", "akram@example.com")

	// Store 2 expired tokens and 1 valid
	_ = rs.StoreToken(ctx, userID, "expired1", time.Now().Add(-1*time.Hour))
	_ = rs.StoreToken(ctx, userID, "expired2", time.Now().Add(-2*time.Hour))
	_ = rs.StoreToken(ctx, userID, "valid1", time.Now().Add(24*time.Hour))

	count, err := rs.DeleteExpired(ctx)
	if err != nil {
		t.Fatalf("DeleteExpired: %v", err)
	}
	if count != 2 {
		t.Errorf("deleted %d, want 2", count)
	}

	// Valid token should still exist
	_, _, _, err = rs.GetByHash(ctx, "valid1")
	if err != nil {
		t.Fatalf("valid token should still exist: %v", err)
	}
}
