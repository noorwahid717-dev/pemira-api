package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"pemira-api/internal/auth"
	"pemira-api/internal/http/response"
	"pemira-api/internal/shared/constants"
	"pemira-api/internal/shared/ctxkeys"
)

// JWTAuth middleware validates JWT token and adds user info to context
func JWTAuth(jwtManager *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Unauthorized(w, "UNAUTHORIZED", "Missing authorization header.")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				response.Unauthorized(w, "UNAUTHORIZED", "Invalid authorization header format.")
				return
			}

			tokenString := parts[1]
			claims, err := jwtManager.ValidateAccessToken(tokenString)
			if err != nil {
				if errors.Is(err, auth.ErrExpiredToken) {
					response.Unauthorized(w, "TOKEN_EXPIRED", "Token sudah kadaluarsa.")
				} else {
					response.Unauthorized(w, "INVALID_TOKEN", "Token tidak valid.")
				}
				return
			}

			// Add user info to context
			ctx := context.WithValue(r.Context(), ctxkeys.UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, ctxkeys.UserRoleKey, string(claims.Role))
			
			if claims.VoterID != nil {
				ctx = context.WithValue(ctx, ctxkeys.VoterIDKey, *claims.VoterID)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuthStudentOnly ensures only STUDENT role can access
func AuthStudentOnly(jwtManager *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return JWTAuth(jwtManager)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := ctxkeys.GetUserRole(r.Context())
			if !ok || role != string(constants.RoleStudent) {
				response.Forbidden(w, "FORBIDDEN", "Akses ditolak. Hanya untuk mahasiswa.")
				return
			}
			next.ServeHTTP(w, r)
		}))
	}
}

// AuthAdminOnly ensures only ADMIN or SUPER_ADMIN role can access
func AuthAdminOnly(jwtManager *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return JWTAuth(jwtManager)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := ctxkeys.GetUserRole(r.Context())
			if !ok {
				response.Forbidden(w, "FORBIDDEN", "Akses ditolak.")
				return
			}
			
			if role != string(constants.RoleAdmin) && role != string(constants.RoleSuperAdmin) {
				response.Forbidden(w, "FORBIDDEN", "Akses ditolak. Hanya untuk admin.")
				return
			}
			
			next.ServeHTTP(w, r)
		}))
	}
}

// AuthTPSOperatorOnly ensures only TPS_OPERATOR role can access
func AuthTPSOperatorOnly(jwtManager *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return JWTAuth(jwtManager)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := ctxkeys.GetUserRole(r.Context())
			if !ok || role != string(constants.RoleTPSOperator) {
				response.Forbidden(w, "FORBIDDEN", "Akses ditolak. Hanya untuk operator TPS.")
				return
			}
			next.ServeHTTP(w, r)
		}))
	}
}
