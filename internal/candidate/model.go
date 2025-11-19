package candidate

import "time"

// CandidateStatus represents the publication status of a candidate
type CandidateStatus string

const (
CandidateStatusDraft     CandidateStatus = "DRAFT"
CandidateStatusPublished CandidateStatus = "PUBLISHED"
CandidateStatusHidden    CandidateStatus = "HIDDEN"
CandidateStatusArchived  CandidateStatus = "ARCHIVED"
)

// Candidate represents a candidate entity in the domain
type Candidate struct {
ID               int64           `json:"id"`
ElectionID       int64           `json:"election_id"`
Number           int             `json:"number"`
Name             string          `json:"name"`
PhotoURL         string          `json:"photo_url"`
ShortBio         string          `json:"short_bio"`
LongBio          string          `json:"long_bio"`
Tagline          string          `json:"tagline"`
FacultyName      string          `json:"faculty_name"`
StudyProgramName string          `json:"study_program_name"`
CohortYear       *int            `json:"cohort_year,omitempty"`

Vision       string        `json:"vision"`
Missions     []string      `json:"missions"`
MainPrograms []MainProgram `json:"main_programs"`
Media        Media         `json:"media"`
SocialLinks  []SocialLink  `json:"social_links"`

Status CandidateStatus `json:"status"`

CreatedAt time.Time `json:"created_at"`
UpdatedAt time.Time `json:"updated_at"`
}

// MainProgram represents a main program of a candidate
type MainProgram struct {
Title       string `json:"title"`
Description string `json:"description"`
Category    string `json:"category"`
}

// Media represents candidate media assets
type Media struct {
VideoURL             *string  `json:"video_url,omitempty"`
GalleryPhotos        []string `json:"gallery_photos,omitempty"`
DocumentManifestoURL *string  `json:"document_manifesto_url,omitempty"`
}

// SocialLink represents a social media link
type SocialLink struct {
Platform string `json:"platform"` // "instagram", "tiktok", etc.
URL      string `json:"url"`
}

// CandidateStats represents voting statistics for a candidate
type CandidateStats struct {
TotalVotes int64   `json:"total_votes"`
Percentage float64 `json:"percentage"`
}
