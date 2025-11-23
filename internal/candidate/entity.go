package candidate

import "time"

type Candidate struct {
	ID               int64                `json:"id"`
	ElectionID       int64                `json:"election_id"`
	Number           int                  `json:"number"`
	Name             string               `json:"name"`
	PhotoURL         string               `json:"photo_url"`
	PhotoMediaID     *string              `json:"photo_media_id,omitempty"`
	ShortBio         string               `json:"short_bio"`
	LongBio          string               `json:"long_bio"`
	Tagline          string               `json:"tagline"`
	FacultyName      string               `json:"faculty_name"`
	StudyProgramName string               `json:"study_program_name"`
	CohortYear       *int                 `json:"cohort_year,omitempty"`
	Vision           string               `json:"vision"`
	Missions         []string             `json:"missions"`
	MainPrograms     []MainProgram        `json:"main_programs"`
	Media            Media                `json:"media"`
	MediaFiles       []CandidateMediaMeta `json:"media_files,omitempty"`
	SocialLinks      []SocialLink         `json:"social_links"`
	Status           CandidateStatus      `json:"status"`
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
}

type CandidateMember struct {
	ID          int64  `json:"id"`
	CandidateID int64  `json:"candidate_id"`
	Name        string `json:"name"`
	Position    string `json:"position"`
	PhotoURL    string `json:"photo_url"`
}
