package client_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/tui/client"
)

func TestGetTimeline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/timeline" {
			t.Errorf("path = %q, want /api/v1/timeline", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("missing auth header")
		}
		json.NewEncoder(w).Encode(models.TimelineResponse{
			Posts:   []models.Post{{ID: "1", Content: "Hello"}},
			HasMore: false,
		})
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	c.SetToken("test-token")

	resp, err := c.GetTimeline("", 50)
	if err != nil {
		t.Fatalf("GetTimeline: %v", err)
	}
	if len(resp.Posts) != 1 {
		t.Errorf("got %d posts, want 1", len(resp.Posts))
	}
}

func TestCreatePost(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/v1/posts" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"post": models.Post{ID: "new-1", Content: "Test post"},
		})
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	c.SetToken("test-token")

	post, err := c.CreatePost("Test post")
	if err != nil {
		t.Fatalf("CreatePost: %v", err)
	}
	if post.Content != "Test post" {
		t.Errorf("content = %q, want %q", post.Content, "Test post")
	}
}

func TestAuthRefreshOn401(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if r.URL.Path == "/api/v1/auth/refresh" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"tokens": models.TokenPair{
					AccessToken:  "new-access",
					RefreshToken: "new-refresh",
				},
			})
			return
		}
		if callCount == 1 {
			// First call: return 401
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]string{"code": "token_expired"},
			})
			return
		}
		// Retry after refresh: success
		json.NewEncoder(w).Encode(models.TimelineResponse{
			Posts: []models.Post{{ID: "1", Content: "After refresh"}},
		})
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	c.SetToken("expired-token")
	c.SetRefreshToken("valid-refresh")

	resp, err := c.GetTimeline("", 50)
	if err != nil {
		t.Fatalf("GetTimeline with refresh: %v", err)
	}
	if len(resp.Posts) != 1 {
		t.Errorf("got %d posts after refresh, want 1", len(resp.Posts))
	}
}
