package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered", "err", err, "stack", string(debug.Stack()))
				writeError(w, http.StatusInternalServerError, "internal_error", "something went wrong, please try again")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
