package voting

import "errors"

var (
	ErrElectionNotFound      = errors.New("election not found")
	ErrElectionNotOpen       = errors.New("election is not open for voting")
	ErrNotEligible           = errors.New("voter is not eligible")
	ErrAlreadyVoted          = errors.New("voter has already voted")
	ErrDuplicateVoteAttempt  = errors.New("duplicate vote attempt")
	ErrCandidateNotFound     = errors.New("candidate not found")
	ErrCandidateInactive     = errors.New("candidate is not active")
	ErrMethodNotAllowed      = errors.New("voting method not allowed")
	ErrNotTPSVoter           = errors.New("voter is not allowed to vote via TPS")
	ErrTPSCheckinNotFound    = errors.New("TPS check-in not found")
	ErrTPSCheckinNotApproved = errors.New("TPS check-in not approved")
	ErrNoActiveCheckin       = errors.New("no active TPS check-in")
	ErrCheckinExpired        = errors.New("TPS check-in has expired")
	ErrTPSNotFound           = errors.New("TPS not found")
	ErrVoterMappingMissing   = errors.New("voter mapping missing")
	ErrInvalidBallotQR       = errors.New("invalid ballot qr")
	ErrElectionMismatch      = errors.New("election mismatch")
	ErrModeNotAllowed        = errors.New("voting mode not available")
)

func translateNotFound(err error, customErr error) error {
	if err != nil && err.Error() == "no rows in result set" {
		return customErr
	}
	return err
}
