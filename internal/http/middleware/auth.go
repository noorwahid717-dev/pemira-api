package middleware

import (
	"context"
	"net/http"
	"strings"

	"pemira-api/internal/auth"
	"pemira-api/internal/http/response"
	"pemira-api/internal/shared/ctxkeys"
)

func Auth(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Unauthorized(w, "Missing authorization header")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				response.Unauthorized(w, "Invalid authorization header")
				return
			}

			claims, err := authService.VerifyToken(parts[1])
			if err != nil {
				response.Unauthorized(w, "Invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), ctxkeys.UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, ctxkeys.UserRoleKey, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
