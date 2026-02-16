package service_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/server/service"
)

func TestCreatePost(t *testing.T) {
	postStore := newMockPostStore()
	svc := service.NewPostService(postStore)

	post, err := svc.CreatePost(context.Background(), "user-123", "Hello, Niotebook!")
	if err != nil {
		t.Fatalf("CreatePost: %v", err)
	}
	if post.Content != "Hello, Niotebook!" {
		t.Errorf("content = %q, want %q", post.Content, "Hello, Niotebook!")
	}
}

func TestCreatePostTrimmed(t *testing.T) {
	postStore := newMockPostStore()
	svc := service.NewPostService(postStore)

	post, _ := svc.CreatePost(context.Background(), "user-123", "  Hello  ")
	if post.Content != "Hello" {
		t.Errorf("content = %q, want trimmed %q", post.Content, "Hello")
	}
}

func TestCreatePostTooLong(t *testing.T) {
	postStore := newMockPostStore()
	svc := service.NewPostService(postStore)

	_, err := svc.CreatePost(context.Background(), "user-123", strings.Repeat("a", 141))
	if err == nil {
		t.Fatal("expected error for too-long post")
	}
}

func TestCreatePostEmpty(t *testing.T) {
	postStore := newMockPostStore()
	svc := service.NewPostService(postStore)

	_, err := svc.CreatePost(context.Background(), "user-123", "   ")
	if err == nil {
		t.Fatal("expected error for empty post")
	}
}

func TestGetTimeline(t *testing.T) {
	postStore := newMockPostStore()
	svc := service.NewPostService(postStore)

	// Add posts via mock
	postStore.AddPost("1", "user-1", "First", time.Now().Add(-2*time.Minute))
	postStore.AddPost("2", "user-1", "Second", time.Now().Add(-1*time.Minute))

	posts, err := svc.GetTimeline(context.Background(), time.Now(), 50)
	if err != nil {
		t.Fatalf("GetTimeline: %v", err)
	}
	if len(posts) != 2 {
		t.Errorf("got %d posts, want 2", len(posts))
	}
}
