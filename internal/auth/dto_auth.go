package auth

// LoginRequest represents login request payload
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents login response
type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	User         *AuthUser `json:"user"`
}

// RefreshRequest represents refresh token request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshResponse represents refresh token response
type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// LogoutRequest represents logout request
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// MeResponse represents current user response
type MeResponse struct {
	*AuthUser
}

// RegisterStudentRequest represents student registration payload
type RegisterStudentRequest struct {
	NIM              string `json:"nim"`
	Name             string `json:"name"`
	Email            string `json:"email,omitempty"`
	FacultyName      string `json:"faculty_name"`
	StudyProgramName string `json:"study_program_name"`
	Semester         string `json:"semester"`
	Password         string `json:"password"`
	VotingMode       string `json:"voting_mode,omitempty"` // ONLINE | TPS
}

// RegisterLecturerStaffRequest represents lecturer/staff registration payload
type RegisterLecturerStaffRequest struct {
	Type           string `json:"type"` // "LECTURER" or "STAFF"
	NIDN           string `json:"nidn,omitempty"`
	NIP            string `json:"nip,omitempty"`
	Name           string `json:"name"`
	Email          string `json:"email,omitempty"`
	FacultyName    string `json:"faculty_name,omitempty"`
	DepartmentName string `json:"department_name,omitempty"`
	UnitName       string `json:"unit_name,omitempty"`
	Position       string `json:"position,omitempty"`
	Password       string `json:"password"`
	VotingMode     string `json:"voting_mode,omitempty"` // ONLINE | TPS (optional)
}

// VoterRegistration contains data needed to create a voter row
type VoterRegistration struct {
	NIM              string
	Name             string
	Email            string
	FacultyName      string
	StudyProgramName string
	Semester         string
}

// LecturerRegistration contains data needed to create a lecturer row
type LecturerRegistration struct {
	NIDN           string
	Name           string
	Email          string
	FacultyName    string
	DepartmentName string
	Position       string
}

// StaffRegistration contains data needed to create a staff row
type StaffRegistration struct {
	NIP      string
	Name     string
	Email    string
	UnitName string
	Position string
	Status   string
}
