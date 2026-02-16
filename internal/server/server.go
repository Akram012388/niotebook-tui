package server

import (
	"net/http"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/server/handler"
	"github.com/Akram012388/niotebook-tui/internal/server/middleware"
	"github.com/Akram012388/niotebook-tui/internal/server/service"
	"github.com/Akram012388/niotebook-tui/internal/server/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	JWTSecret  string
	Host       string
	Port       string
	CORSOrigin string
}

func NewServer(cfg *Config, pool *pgxpool.Pool) *http.Server {
	// Stores
	userStore := store.NewUserStore(pool)
	postStore := store.NewPostStore(pool)
	tokenStore := store.NewRefreshTokenStore(pool)

	// Services
	authSvc := service.NewAuthService(userStore, tokenStore, cfg.JWTSecret)
	postSvc := service.NewPostService(postStore)
	userSvc := service.NewUserService(userStore)

	// Router (Go 1.22 pattern matching)
	mux := http.NewServeMux()

	// Auth routes
	mux.HandleFunc("POST /api/v1/auth/register", handler.HandleRegister(authSvc))
	mux.HandleFunc("POST /api/v1/auth/login", handler.HandleLogin(authSvc))
	mux.HandleFunc("POST /api/v1/auth/refresh", handler.HandleRefresh(authSvc))

	// Post routes
	mux.HandleFunc("POST /api/v1/posts", handler.HandleCreatePost(postSvc))
	mux.HandleFunc("GET /api/v1/posts/{id}", handler.HandleGetPost(postSvc))

	// Timeline
	mux.HandleFunc("GET /api/v1/timeline", handler.HandleTimeline(postSvc))

	// User routes
	mux.HandleFunc("GET /api/v1/users/{id}", handler.HandleGetUser(userSvc))
	mux.HandleFunc("GET /api/v1/users/{id}/posts", handler.HandleGetUserPosts(postSvc))
	mux.HandleFunc("PATCH /api/v1/users/me", handler.HandleUpdateUser(userSvc))

	// Health
	mux.HandleFunc("GET /health", handler.HandleHealth(pool))

	// Middleware chain: Recovery → Logging → RateLimit → CORS → Auth → Handler
	rateLimiter := middleware.NewRateLimiter()
	var h http.Handler = mux
	h = middleware.Auth(cfg.JWTSecret)(h)
	h = middleware.CORS(cfg.CORSOrigin)(h)
	h = rateLimiter.Middleware(h)
	h = middleware.Logging(h)
	h = middleware.Recovery(h)

	return &http.Server{
		Addr:         cfg.Host + ":" + cfg.Port,
		Handler:      h,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}
