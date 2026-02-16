package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimiter_AuthEndpoint(t *testing.T) {
	rl := NewRateLimiter()
	defer rl.Stop()

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Auth endpoints have burst of 5. Send 12 rapid requests; first 5 should pass,
	// remaining should be rate limited.
	var passed, limited int
	for i := 0; i < 12; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		switch rr.Code {
		case http.StatusOK:
			passed++
		case http.StatusTooManyRequests:
			limited++
			if rr.Header().Get("Retry-After") == "" {
				t.Error("expected Retry-After header on 429 response")
			}
		default:
			t.Errorf("unexpected status code: %d", rr.Code)
		}
	}

	if passed == 0 {
		t.Error("expected at least some requests to pass")
	}
	if limited == 0 {
		t.Error("expected at least some requests to be rate limited")
	}
	// With burst 5, exactly 5 should pass and 7 should be limited
	if passed != 5 {
		t.Errorf("expected 5 passed requests, got %d", passed)
	}
	if limited != 7 {
		t.Errorf("expected 7 rate limited requests, got %d", limited)
	}
	t.Logf("passed=%d limited=%d", passed, limited)
}

func TestRateLimiter_HealthExempt(t *testing.T) {
	rl := NewRateLimiter()
	defer rl.Stop()

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// /health should be exempt from rate limiting
	for i := 0; i < 50; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i, rr.Code)
		}
	}
}

func TestRateLimiter_WriteEndpoint(t *testing.T) {
	rl := NewRateLimiter()
	defer rl.Stop()

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	// Write endpoint: POST /api/v1/posts — burst 10
	var passed, limited int
	for i := 0; i < 15; i++ {
		req := httptest.NewRequest("POST", "/api/v1/posts", nil)
		req.RemoteAddr = "10.0.0.1:5000"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		switch rec.Code {
		case http.StatusCreated:
			passed++
		case http.StatusTooManyRequests:
			limited++
		default:
			t.Errorf("unexpected status: %d", rec.Code)
		}
	}

	if passed == 0 {
		t.Error("expected at least some write requests to pass")
	}
	if limited == 0 {
		t.Error("expected at least some write requests to be rate limited")
	}
}

func TestRateLimiter_ReadEndpoint(t *testing.T) {
	rl := NewRateLimiter()
	defer rl.Stop()

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Read endpoint: GET /api/v1/timeline — burst 30
	req := httptest.NewRequest("GET", "/api/v1/timeline", nil)
	req.RemoteAddr = "10.0.0.1:5000"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestRateLimiter_DifferentIPsIndependent(t *testing.T) {
	rl := NewRateLimiter()
	defer rl.Stop()

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Exhaust limits for IP 1
	for i := 0; i < 12; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	// IP 2 should still be able to make requests
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.RemoteAddr = "10.0.0.2:12345"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("different IP should not be rate limited, got %d", rr.Code)
	}
}
