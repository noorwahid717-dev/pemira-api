package shared

import "errors"

var (
	ErrNotFound          = errors.New("resource not found")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrBadRequest        = errors.New("bad request")
	ErrConflict          = errors.New("conflict")
	ErrInternalServer    = errors.New("internal server error")
	
	ErrAlreadyVoted      = errors.New("already voted")
	ErrInvalidPhase      = errors.New("invalid election phase")
	ErrInvalidVotingMode = errors.New("invalid voting mode")
	ErrVoterNotEligible  = errors.New("voter not eligible")
	ErrInvalidToken      = errors.New("invalid token")
	ErrDuplicateEntry    = errors.New("duplicate entry")
)
