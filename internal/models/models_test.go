package models_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/models"
)

func TestUserJSONRoundTrip(t *testing.T) {
	user := models.User{
		ID:          "550e8400-e29b-41d4-a716-446655440000",
		Username:    "akram",
		DisplayName: "Akram",
		Bio:         "Building things in Go.",
		CreatedAt:   time.Date(2026, 2, 15, 22, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got models.User
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.ID != user.ID || got.Username != user.Username || got.DisplayName != user.DisplayName {
		t.Errorf("round trip mismatch: got %+v, want %+v", got, user)
	}
}

func TestPostJSONRoundTrip(t *testing.T) {
	post := models.Post{
		ID:       "660e8400-e29b-41d4-a716-446655440001",
		AuthorID: "550e8400-e29b-41d4-a716-446655440000",
		Content:  "Hello, Niotebook!",
		CreatedAt: time.Date(2026, 2, 15, 23, 30, 0, 0, time.UTC),
	}

	data, err := json.Marshal(post)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got models.Post
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.ID != post.ID || got.Content != post.Content {
		t.Errorf("round trip mismatch: got %+v, want %+v", got, post)
	}
}

func TestPostJSONIncludesAuthor(t *testing.T) {
	post := models.Post{
		ID:       "660e8400-e29b-41d4-a716-446655440001",
		AuthorID: "550e8400-e29b-41d4-a716-446655440000",
		Author: &models.User{
			ID:       "550e8400-e29b-41d4-a716-446655440000",
			Username: "akram",
		},
		Content:   "Hello!",
		CreatedAt: time.Date(2026, 2, 15, 23, 30, 0, 0, time.UTC),
	}

	data, err := json.Marshal(post)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal raw: %v", err)
	}

	if _, ok := raw["author"]; !ok {
		t.Error("expected author field in JSON")
	}
}

func TestPostJSONOmitsNilAuthor(t *testing.T) {
	post := models.Post{
		ID:        "660e8400-e29b-41d4-a716-446655440001",
		AuthorID:  "550e8400-e29b-41d4-a716-446655440000",
		Content:   "Hello!",
		CreatedAt: time.Date(2026, 2, 15, 23, 30, 0, 0, time.UTC),
	}

	data, err := json.Marshal(post)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal raw: %v", err)
	}

	if _, ok := raw["author"]; ok {
		t.Error("expected author field to be omitted when nil")
	}
}

func TestAPIErrorJSON(t *testing.T) {
	apiErr := models.APIError{
		Code:    "validation_error",
		Message: "Username must be 3-15 characters",
		Field:   "username",
	}

	data, err := json.Marshal(apiErr)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got models.APIError
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Code != apiErr.Code || got.Field != apiErr.Field {
		t.Errorf("round trip mismatch: got %+v, want %+v", got, apiErr)
	}
}
