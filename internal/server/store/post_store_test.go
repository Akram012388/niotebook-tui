package store_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/server/store"
)

func createTestUser(t *testing.T, s store.UserStore, username, email string) string {
	t.Helper()
	user, err := s.CreateUser(context.Background(), username, email, "$2a$12$fakehash", username)
	if err != nil {
		t.Fatalf("create test user: %v", err)
	}
	return user.ID
}

func TestCreatePost(t *testing.T) {
	pool := setupTestDB(t)
	us := store.NewUserStore(pool)
	ps := store.NewPostStore(pool)
	ctx := context.Background()

	userID := createTestUser(t, us, "akram", "akram@example.com")

	post, err := ps.CreatePost(ctx, userID, "Hello, Niotebook!")
	if err != nil {
		t.Fatalf("CreatePost: %v", err)
	}
	if post.ID == "" {
		t.Error("expected non-empty ID")
	}
	if post.Content != "Hello, Niotebook!" {
		t.Errorf("content = %q, want %q", post.Content, "Hello, Niotebook!")
	}
	if post.AuthorID != userID {
		t.Errorf("author_id = %q, want %q", post.AuthorID, userID)
	}
}

func TestCreatePostContentTooLong(t *testing.T) {
	pool := setupTestDB(t)
	us := store.NewUserStore(pool)
	ps := store.NewPostStore(pool)
	ctx := context.Background()

	userID := createTestUser(t, us, "akram", "akram@example.com")

	_, err := ps.CreatePost(ctx, userID, strings.Repeat("a", 141))
	if err == nil {
		t.Fatal("expected error for content too long")
	}
}

func TestCreatePostEmpty(t *testing.T) {
	pool := setupTestDB(t)
	us := store.NewUserStore(pool)
	ps := store.NewPostStore(pool)
	ctx := context.Background()

	userID := createTestUser(t, us, "akram", "akram@example.com")

	_, err := ps.CreatePost(ctx, userID, "   ")
	if err == nil {
		t.Fatal("expected error for whitespace-only content")
	}
}

func TestGetPostByID(t *testing.T) {
	pool := setupTestDB(t)
	us := store.NewUserStore(pool)
	ps := store.NewPostStore(pool)
	ctx := context.Background()

	userID := createTestUser(t, us, "akram", "akram@example.com")

	created, err := ps.CreatePost(ctx, userID, "Test post")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	post, err := ps.GetPostByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetPostByID: %v", err)
	}
	if post.Content != "Test post" {
		t.Errorf("content = %q, want %q", post.Content, "Test post")
	}
	if post.Author == nil {
		t.Fatal("expected author to be joined")
	}
	if post.Author.Username != "akram" {
		t.Errorf("author.username = %q, want %q", post.Author.Username, "akram")
	}
}

func TestGetPostByIDNotFound(t *testing.T) {
	pool := setupTestDB(t)
	ps := store.NewPostStore(pool)
	ctx := context.Background()

	_, err := ps.GetPostByID(ctx, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatal("expected error for nonexistent post")
	}
}

func TestGetTimeline(t *testing.T) {
	pool := setupTestDB(t)
	us := store.NewUserStore(pool)
	ps := store.NewPostStore(pool)
	ctx := context.Background()

	userID := createTestUser(t, us, "akram", "akram@example.com")

	_, _ = ps.CreatePost(ctx, userID, "First post")
	time.Sleep(10 * time.Millisecond)
	_, _ = ps.CreatePost(ctx, userID, "Second post")
	time.Sleep(10 * time.Millisecond)
	_, _ = ps.CreatePost(ctx, userID, "Third post")

	posts, err := ps.GetTimeline(ctx, time.Now(), 50)
	if err != nil {
		t.Fatalf("GetTimeline: %v", err)
	}
	if len(posts) != 3 {
		t.Fatalf("got %d posts, want 3", len(posts))
	}
	// Should be reverse chronological
	if posts[0].Content != "Third post" {
		t.Errorf("first post = %q, want %q", posts[0].Content, "Third post")
	}
	if posts[2].Content != "First post" {
		t.Errorf("last post = %q, want %q", posts[2].Content, "First post")
	}
}

func TestGetTimelineCursorPagination(t *testing.T) {
	pool := setupTestDB(t)
	us := store.NewUserStore(pool)
	ps := store.NewPostStore(pool)
	ctx := context.Background()

	userID := createTestUser(t, us, "akram", "akram@example.com")

	for i := 0; i < 5; i++ {
		_, _ = ps.CreatePost(ctx, userID, "Post "+string(rune('A'+i)))
		time.Sleep(10 * time.Millisecond)
	}

	// Fetch first page
	page1, err := ps.GetTimeline(ctx, time.Now(), 2)
	if err != nil {
		t.Fatalf("page 1: %v", err)
	}
	if len(page1) != 2 {
		t.Fatalf("page 1 got %d posts, want 2", len(page1))
	}

	// Use the last post's created_at as cursor for next page
	cursor := page1[len(page1)-1].CreatedAt
	page2, err := ps.GetTimeline(ctx, cursor, 2)
	if err != nil {
		t.Fatalf("page 2: %v", err)
	}
	if len(page2) != 2 {
		t.Fatalf("page 2 got %d posts, want 2", len(page2))
	}

	// Pages should not overlap
	if page1[0].ID == page2[0].ID {
		t.Error("pages overlap — same first post")
	}
}

func TestGetUserPosts(t *testing.T) {
	pool := setupTestDB(t)
	us := store.NewUserStore(pool)
	ps := store.NewPostStore(pool)
	ctx := context.Background()

	user1ID := createTestUser(t, us, "akram", "akram@example.com")
	user2ID := createTestUser(t, us, "sara", "sara@example.com")

	_, _ = ps.CreatePost(ctx, user1ID, "Akram's post 1")
	_, _ = ps.CreatePost(ctx, user1ID, "Akram's post 2")
	_, _ = ps.CreatePost(ctx, user2ID, "Sara's post")

	posts, err := ps.GetUserPosts(ctx, user1ID, time.Now(), 50)
	if err != nil {
		t.Fatalf("GetUserPosts: %v", err)
	}
	if len(posts) != 2 {
		t.Fatalf("got %d posts, want 2", len(posts))
	}
	for _, p := range posts {
		if p.AuthorID != user1ID {
			t.Errorf("got post from author %q, want %q", p.AuthorID, user1ID)
		}
	}
}
