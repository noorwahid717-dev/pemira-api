package middleware

import (
	"net/http"

	"pemira-api/internal/http/response"
	"pemira-api/internal/shared/constants"
	"pemira-api/internal/shared/ctxkeys"
)

func RequireRole(roles ...constants.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value(ctxkeys.UserRoleKey).(constants.Role)
			if !ok {
				response.Forbidden(w, "FORBIDDEN", "No role found")
				return
			}

			for _, role := range roles {
				if userRole == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			response.Forbidden(w, "FORBIDDEN", "Insufficient permissions")
		})
	}
}

func RequireAdmin() func(http.Handler) http.Handler {
	return RequireRole(constants.RoleAdmin, constants.RoleSuperAdmin)
}

func RequireTPSOperator() func(http.Handler) http.Handler {
	return RequireRole(constants.RoleTPSOperator, constants.RoleAdmin, constants.RoleSuperAdmin)
}

func RequireStudent() func(http.Handler) http.Handler {
	return RequireRole(constants.RoleStudent)
}
