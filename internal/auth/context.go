package auth

import (
	"context"

	"pemira-api/internal/shared/constants"
	"pemira-api/internal/shared/ctxkeys"
)

// FromContext builds an AuthUser from context values set by JWT middleware.
func FromContext(ctx context.Context) (AuthUser, bool) {
	userID, ok := ctxkeys.GetUserID(ctx)
	if !ok {
		return AuthUser{}, false
	}

	roleStr, ok := ctxkeys.GetUserRole(ctx)
	if !ok {
		return AuthUser{}, false
	}

	var voterID *int64
	if id, ok := ctxkeys.GetVoterID(ctx); ok {
		voterID = &id
	}

	authUser := AuthUser{
		ID:      userID,
		Role:    constants.Role(roleStr),
		VoterID: voterID,
	}

	return authUser, true
}
