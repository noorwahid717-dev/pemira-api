package ctxkeys

import "context"

type contextKey string

const (
	UserIDKey     contextKey = "user_id"
	UserRoleKey   contextKey = "user_role"
	VoterIDKey    contextKey = "voter_id"
	RequestIDKey  contextKey = "request_id"
	ElectionIDKey contextKey = "election_id"
)

// GetVoterID extracts voter ID from context
func GetVoterID(ctx context.Context) (int64, bool) {
	v := ctx.Value(VoterIDKey)
	if v == nil {
		// Fallback to UserIDKey for backward compatibility
		v = ctx.Value(UserIDKey)
		if v == nil {
			return 0, false
		}
	}
	id, ok := v.(int64)
	return id, ok
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) (int64, bool) {
	v := ctx.Value(UserIDKey)
	if v == nil {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}

// GetUserRole extracts user role from context
func GetUserRole(ctx context.Context) (string, bool) {
	v := ctx.Value(UserRoleKey)
	if v == nil {
		return "", false
	}
	role, ok := v.(string)
	return role, ok
}
