package settings

import "time"

type AppSetting struct {
	Key         string     `json:"key" db:"key"`
	Value       string     `json:"value" db:"value"`
	Description string     `json:"description" db:"description"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	UpdatedBy   *int64     `json:"updated_by" db:"updated_by"`
}

type UpdateSettingRequest struct {
	Value string `json:"value"`
}

type SettingsResponse struct {
	ActiveElectionID  int `json:"active_election_id"`
	DefaultElectionID int `json:"default_election_id"`
}
