package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Akram012388/niotebook-tui/internal/server/middleware"
)

func TestRecoveryMiddlewareCatchesPanic(t *testing.T) {
	panicker := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	handler := middleware.Recovery(panicker)
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	errObj, ok := body["error"].(map[string]interface{})
	if !ok {
		t.Fatal("expected error object in response")
	}
	if errObj["code"] != "internal_error" {
		t.Errorf("error code = %q, want %q", errObj["code"], "internal_error")
	}
}

func TestRecoveryMiddlewarePassesThrough(t *testing.T) {
	normal := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Recovery(normal)
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
