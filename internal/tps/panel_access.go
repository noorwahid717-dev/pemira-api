package tps

import "pemira-api/internal/shared/constants"

func hasPanelAccess(role string) bool {
	return role == string(constants.RoleAdmin) ||
		role == string(constants.RoleSuperAdmin) ||
		role == string(constants.RoleTPSOperator)
}
