package announcement

import "time"

type Announcement struct {
	ID          int64     `json:"id"`
	ElectionID  int64     `json:"election_id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Type        string    `json:"type"` // INFO, WARNING, SUCCESS, ERROR
	Priority    int       `json:"priority"` // 1-5, higher = more important
	IsPublished bool      `json:"is_published"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedBy   int64     `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
