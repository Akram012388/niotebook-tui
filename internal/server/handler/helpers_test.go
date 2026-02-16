package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Akram012388/niotebook-tui/internal/models"
)

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	writeJSON(rec, http.StatusOK, map[string]string{"key": "value"})

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["key"] != "value" {
		t.Errorf("body key = %q, want %q", body["key"], "value")
	}
}

func TestWriteAPIErrorWithAPIError(t *testing.T) {
	rec := httptest.NewRecorder()
	writeAPIError(rec, &models.APIError{
		Code:    models.ErrCodeValidation,
		Message: "invalid email",
		Field:   "email",
	})

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var body struct {
		Error models.APIError `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Error.Code != models.ErrCodeValidation {
		t.Errorf("error code = %q, want %q", body.Error.Code, models.ErrCodeValidation)
	}
	if body.Error.Field != "email" {
		t.Errorf("error field = %q, want %q", body.Error.Field, "email")
	}
}

func TestWriteAPIErrorWithGenericError(t *testing.T) {
	rec := httptest.NewRecorder()
	writeAPIError(rec, fmt.Errorf("database connection lost"))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	var body struct {
		Error models.APIError `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Error.Code != models.ErrCodeInternal {
		t.Errorf("error code = %q, want %q", body.Error.Code, models.ErrCodeInternal)
	}
}

func TestErrorCodeToHTTPStatus(t *testing.T) {
	tests := []struct {
		code   string
		status int
	}{
		{models.ErrCodeValidation, http.StatusBadRequest},
		{models.ErrCodeContentLong, http.StatusBadRequest},
		{models.ErrCodeUnauthorized, http.StatusUnauthorized},
		{models.ErrCodeTokenExpired, http.StatusUnauthorized},
		{models.ErrCodeForbidden, http.StatusForbidden},
		{models.ErrCodeNotFound, http.StatusNotFound},
		{models.ErrCodeConflict, http.StatusConflict},
		{models.ErrCodeRateLimited, http.StatusTooManyRequests},
		{models.ErrCodeInternal, http.StatusInternalServerError},
		{"unknown_code", http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := errorCodeToHTTPStatus(tt.code)
			if got != tt.status {
				t.Errorf("errorCodeToHTTPStatus(%q) = %d, want %d", tt.code, got, tt.status)
			}
		})
	}
}

func TestDecodeBody(t *testing.T) {
	body := `{"username":"akram"}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))

	var result struct {
		Username string `json:"username"`
	}
	if err := decodeBody(req, &result); err != nil {
		t.Fatalf("decodeBody: %v", err)
	}
	if result.Username != "akram" {
		t.Errorf("username = %q, want %q", result.Username, "akram")
	}
}

func TestDecodeBodyRejectsUnknownFields(t *testing.T) {
	body := `{"username":"akram","extra":"field"}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))

	var result struct {
		Username string `json:"username"`
	}
	if err := decodeBody(req, &result); err == nil {
		t.Error("expected error for unknown fields, got nil")
	}
}

func TestDecodeBodyRejectsOversizedBody(t *testing.T) {
	// 4096 + 1 bytes should fail
	body := strings.Repeat("a", 4097)
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))

	var result map[string]interface{}
	if err := decodeBody(req, &result); err == nil {
		t.Error("expected error for oversized body, got nil")
	}
}
