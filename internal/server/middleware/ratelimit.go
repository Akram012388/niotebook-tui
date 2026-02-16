package middleware

import (
	"fmt"
	"math"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimitCategory determines the rate limit tier for a request.
type RateLimitCategory int

const (
	categoryExempt RateLimitCategory = iota
	categoryAuth
	categoryWrite
	categoryRead
)

type visitor struct {
	limiters map[RateLimitCategory]*rate.Limiter
	lastSeen time.Time
}

// RateLimiter implements per-IP token bucket rate limiting.
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	done     chan struct{}
}

// NewRateLimiter creates a new rate limiter and starts background cleanup.
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		done:     make(chan struct{}),
	}
	go rl.cleanup()
	return rl
}

// Stop terminates the background cleanup goroutine.
func (rl *RateLimiter) Stop() {
	close(rl.done)
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			for ip, v := range rl.visitors {
				if time.Since(v.lastSeen) > 15*time.Minute {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		case <-rl.done:
			return
		}
	}
}

func (rl *RateLimiter) getVisitor(ip string, cat RateLimitCategory) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		v = &visitor{
			limiters: make(map[RateLimitCategory]*rate.Limiter),
		}
		rl.visitors[ip] = v
	}
	v.lastSeen = time.Now()

	limiter, exists := v.limiters[cat]
	if !exists {
		limiter = newLimiterForCategory(cat)
		v.limiters[cat] = limiter
	}
	return limiter
}

func newLimiterForCategory(cat RateLimitCategory) *rate.Limiter {
	switch cat {
	case categoryAuth:
		return rate.NewLimiter(rate.Every(time.Minute/10), 5)
	case categoryWrite:
		return rate.NewLimiter(rate.Every(time.Minute/30), 10)
	case categoryRead:
		return rate.NewLimiter(rate.Every(time.Minute/120), 30)
	default:
		return rate.NewLimiter(rate.Inf, 0)
	}
}

func categorize(r *http.Request) RateLimitCategory {
	path := r.URL.Path

	if path == "/health" {
		return categoryExempt
	}

	if strings.HasPrefix(path, "/api/v1/auth/") {
		return categoryAuth
	}

	if r.Method == http.MethodPost && strings.HasPrefix(path, "/api/v1/posts") {
		return categoryWrite
	}

	return categoryRead
}

func extractIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// Middleware returns an HTTP middleware that enforces rate limits.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cat := categorize(r)
		if cat == categoryExempt {
			next.ServeHTTP(w, r)
			return
		}

		ip := extractIP(r)
		limiter := rl.getVisitor(ip, cat)

		if !limiter.Allow() {
			retryAfter := math.Ceil(float64(time.Second) / float64(limiter.Limit()) / float64(time.Second))
			if retryAfter < 1 {
				retryAfter = 1
			}
			w.Header().Set("Retry-After", fmt.Sprintf("%.0f", retryAfter))
			writeError(w, http.StatusTooManyRequests, "rate_limited", "rate limited, try again later")
			return
		}

		next.ServeHTTP(w, r)
	})
}
