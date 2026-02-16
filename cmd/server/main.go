package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/build"
	"github.com/Akram012388/niotebook-tui/internal/server"
	"github.com/Akram012388/niotebook-tui/internal/server/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	port := flag.String("port", envOrDefault("NIOTEBOOK_PORT", "8080"), "listen port")
	host := flag.String("host", envOrDefault("NIOTEBOOK_HOST", "localhost"), "listen host")
	_ = flag.Bool("migrate", false, "run pending migrations on startup")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		slog.Info("niotebook-server", "version", build.Version, "commit", build.CommitSHA)
		os.Exit(0)
	}

	// Configure logging
	logLevel := slog.LevelInfo
	if lvl := os.Getenv("NIOTEBOOK_LOG_LEVEL"); lvl == "debug" {
		logLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})))

	dbURL := os.Getenv("NIOTEBOOK_DB_URL")
	if dbURL == "" {
		slog.Error("NIOTEBOOK_DB_URL is required")
		os.Exit(1)
	}

	jwtSecret := os.Getenv("NIOTEBOOK_JWT_SECRET")
	if jwtSecret == "" {
		slog.Error("NIOTEBOOK_JWT_SECRET is required")
		os.Exit(1)
	}
	if len(jwtSecret) < 32 {
		slog.Error("NIOTEBOOK_JWT_SECRET must be at least 32 bytes", "length", len(jwtSecret))
		os.Exit(1)
	}

	corsOrigin := envOrDefault("NIOTEBOOK_CORS_ORIGIN", "*")

	// Database
	ctx := context.Background()
	pool, err := store.NewPool(ctx, dbURL)
	if err != nil {
		slog.Error("database connection failed", "err", err)
		os.Exit(1)
	}

	// Server
	cfg := &server.Config{JWTSecret: jwtSecret, Host: *host, Port: *port, CORSOrigin: corsOrigin}
	srv := server.NewServer(cfg, pool)

	go func() {
		slog.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	// Background: token cleanup
	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
	go runTokenCleanup(cleanupCtx, pool)

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")
	cleanupCancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("forced shutdown", "err", err)
	}

	pool.Close()
	slog.Info("server stopped")
}

func runTokenCleanup(ctx context.Context, pool *pgxpool.Pool) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, err := pool.Exec(ctx, "DELETE FROM refresh_tokens WHERE expires_at < NOW()")
			if err != nil {
				slog.Error("token cleanup failed", "err", err)
			} else {
				slog.Debug("running token cleanup")
			}
		}
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
