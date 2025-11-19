package candidate

import (
"context"
"errors"
)

// Filter represents query filters for listing candidates
type Filter struct {
Status *CandidateStatus
Search string
Limit  int
Offset int
}

// CandidateRepository defines the interface for candidate data access
type CandidateRepository interface {
// ListByElection returns candidates for an election with pagination
ListByElection(ctx context.Context, electionID int64, filter Filter) ([]Candidate, int64, error)

// GetByID returns a single candidate by election and candidate ID
GetByID(ctx context.Context, electionID, candidateID int64) (*Candidate, error)
}

// Common errors
var (
ErrCandidateNotFound = errors.New("candidate not found")
)
