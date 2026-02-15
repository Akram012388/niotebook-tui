package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/build"
	"github.com/jackc/pgx/v5/pgxpool"
)

func HandleHealth(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"status":  "error",
				"message": "database connection failed",
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"status":  "ok",
			"version": build.Version,
		})
	}
}
