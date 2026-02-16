package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userCtxKey contextKey = "user_claims"

type UserClaims struct {
	UserID   string
	Username string
}

func UserIDFromContext(ctx context.Context) string {
	claims, ok := ctx.Value(userCtxKey).(*UserClaims)
	if !ok {
		return ""
	}
	return claims.UserID
}

func UsernameFromContext(ctx context.Context) string {
	claims, ok := ctx.Value(userCtxKey).(*UserClaims)
	if !ok {
		return ""
	}
	return claims.Username
}

var exemptPaths = map[string]bool{
	"/api/v1/auth/login":    true,
	"/api/v1/auth/register": true,
	"/api/v1/auth/refresh":  true,
	"/health":               true,
}

func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if exemptPaths[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				writeError(w, http.StatusUnauthorized, models.ErrCodeUnauthorized, "Missing or invalid authorization header")
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			}, jwt.WithValidMethods([]string{"HS256"}))

			if err != nil || !token.Valid {
				code := models.ErrCodeUnauthorized
				if strings.Contains(err.Error(), "expired") {
					code = models.ErrCodeTokenExpired
				}
				writeError(w, http.StatusUnauthorized, code, "Invalid or expired token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				writeError(w, http.StatusUnauthorized, models.ErrCodeUnauthorized, "Invalid token claims")
				return
			}

			// Safe type assertions — return 401 instead of panicking
			sub, ok := claims["sub"].(string)
			if !ok || sub == "" {
				writeError(w, http.StatusUnauthorized, models.ErrCodeUnauthorized, "invalid token claims")
				return
			}
			uname, ok := claims["username"].(string)
			if !ok {
				writeError(w, http.StatusUnauthorized, models.ErrCodeUnauthorized, "invalid token claims")
				return
			}

			userClaims := &UserClaims{
				UserID:   sub,
				Username: uname,
			}

			ctx := context.WithValue(r.Context(), userCtxKey, userClaims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
