package announcement

type CreateAnnouncementRequest struct {
	ElectionID  int64  `json:"election_id" validate:"required"`
	Title       string `json:"title" validate:"required"`
	Content     string `json:"content" validate:"required"`
	Type        string `json:"type" validate:"required,oneof=INFO WARNING SUCCESS ERROR"`
	Priority    int    `json:"priority" validate:"min=1,max=5"`
	IsPublished bool   `json:"is_published"`
}

type UpdateAnnouncementRequest struct {
	Title       string `json:"title"`
	Content     string `json:"content"`
	Type        string `json:"type" validate:"oneof=INFO WARNING SUCCESS ERROR"`
	Priority    int    `json:"priority" validate:"min=1,max=5"`
	IsPublished bool   `json:"is_published"`
}
