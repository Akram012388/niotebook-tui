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
		_ = json.NewEncoder(w).Encode(models.TimelineResponse{
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
		_ = json.NewEncoder(w).Encode(map[string]any{
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

func TestClientTimeout(t *testing.T) {
	c := client.New("http://198.51.100.1:1") // Non-routable IP — will timeout
	c.SetToken("test-token")

	_, err := c.GetTimeline("", 20)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestAuthRefreshOn401(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if r.URL.Path == "/api/v1/auth/refresh" {
			_ = json.NewEncoder(w).Encode(map[string]any{
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
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]string{"code": "token_expired"},
			})
			return
		}
		// Retry after refresh: success
		_ = json.NewEncoder(w).Encode(models.TimelineResponse{
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

func TestLogin(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/v1/auth/login" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(models.AuthResponse{
			User:   &models.User{ID: "u1", Username: "akram"},
			Tokens: &models.TokenPair{AccessToken: "at", RefreshToken: "rt"},
		})
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	resp, err := c.Login("akram@example.com", "password123")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if resp.User.Username != "akram" {
		t.Errorf("username = %q, want %q", resp.User.Username, "akram")
	}
}

func TestRegister(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/v1/auth/register" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(models.AuthResponse{
			User:   &models.User{ID: "u1", Username: "akram"},
			Tokens: &models.TokenPair{AccessToken: "at", RefreshToken: "rt"},
		})
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	resp, err := c.Register("akram", "akram@example.com", "password123")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if resp.User.Username != "akram" {
		t.Errorf("username = %q, want %q", resp.User.Username, "akram")
	}
}

func TestGetPost(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/api/v1/posts/post-1" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"post": models.Post{ID: "post-1", Content: "A post"},
		})
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	c.SetToken("test-token")

	post, err := c.GetPost("post-1")
	if err != nil {
		t.Fatalf("GetPost: %v", err)
	}
	if post.ID != "post-1" {
		t.Errorf("ID = %q, want %q", post.ID, "post-1")
	}
}

func TestGetUser(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/api/v1/users/me" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"user": models.User{ID: "u1", Username: "akram"},
		})
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	c.SetToken("test-token")

	user, err := c.GetUser("me")
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if user.Username != "akram" {
		t.Errorf("username = %q, want %q", user.Username, "akram")
	}
}

func TestGetUserPosts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/users/u1/posts" {
			t.Errorf("path = %q, want /api/v1/users/u1/posts", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(models.TimelineResponse{
			Posts:   []models.Post{{ID: "p1", Content: "User post"}},
			HasMore: false,
		})
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	c.SetToken("test-token")

	resp, err := c.GetUserPosts("u1", "", 20)
	if err != nil {
		t.Fatalf("GetUserPosts: %v", err)
	}
	if len(resp.Posts) != 1 {
		t.Errorf("got %d posts, want 1", len(resp.Posts))
	}
}

func TestUpdateUser(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" || r.URL.Path != "/api/v1/users/me" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"user": models.User{ID: "u1", Username: "akram", DisplayName: "New Name"},
		})
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	c.SetToken("test-token")

	name := "New Name"
	user, err := c.UpdateUser(&models.UserUpdate{DisplayName: &name})
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	if user.DisplayName != "New Name" {
		t.Errorf("display_name = %q, want %q", user.DisplayName, "New Name")
	}
}

func TestOnTokenRefresh(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/refresh" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"tokens": models.TokenPair{
					AccessToken:  "new-at",
					RefreshToken: "new-rt",
				},
			})
			return
		}
		_ = json.NewEncoder(w).Encode(models.TimelineResponse{})
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	c.SetRefreshToken("old-rt")

	var gotAT, gotRT string
	c.OnTokenRefresh(func(at, rt string) {
		gotAT = at
		gotRT = rt
	})

	_, err := c.Refresh()
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if gotAT != "new-at" {
		t.Errorf("callback access token = %q, want %q", gotAT, "new-at")
	}
	if gotRT != "new-rt" {
		t.Errorf("callback refresh token = %q, want %q", gotRT, "new-rt")
	}
}

func TestAPIErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "validation_error",
				"message": "content too short",
			},
		})
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	c.SetToken("test-token")

	_, err := c.GetTimeline("", 20)
	if err == nil {
		t.Fatal("expected error for 400 response")
	}
}
