package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/JoX23/go-without-magic/internal/service"
)

type contextKey string

const UserIDKey contextKey = "userID"
const UserEmailKey contextKey = "userEmail"

// JWTAuth returns a middleware that validates Bearer JWT tokens using TokenService.
// It first checks the Authorization header, then falls back to the auth_token cookie.
func JWTAuth(tokenSvc *service.TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tokenStr string

			// 1. Intentar header Authorization
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
			}

			// 2. Fallback a cookie
			if tokenStr == "" {
				if cookie, err := r.Cookie("auth_token"); err == nil {
					tokenStr = cookie.Value
				}
			}

			if tokenStr == "" {
				http.Error(w, `{"error":"missing or invalid authorization header"}`, http.StatusUnauthorized)
				return
			}

			claims, err := tokenSvc.Validate(tokenStr)
			if err != nil {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuth wraps a single handler with JWT auth.
func RequireAuth(tokenSvc *service.TokenService, next http.HandlerFunc) http.HandlerFunc {
	return JWTAuth(tokenSvc)(next).ServeHTTP
}
