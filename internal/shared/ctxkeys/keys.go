package ctxkeys

type contextKey string

const (
	UserIDKey     contextKey = "user_id"
	UserRoleKey   contextKey = "user_role"
	RequestIDKey  contextKey = "request_id"
	ElectionIDKey contextKey = "election_id"
)
