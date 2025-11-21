package voting

import (
	"strconv"
	"strings"
)

type BallotQR struct {
	ElectionID  int64
	CandidateID int64
	Version     int
}

func parseBallotQR(raw string) (*BallotQR, error) {
	parts := strings.Split(raw, "|")
	if len(parts) < 4 || parts[0] != "PEMIRA-UNIWA" {
		return nil, ErrInvalidBallotQR
	}

	var (
		electionID  int64
		candidateID int64
		version     int
	)

	for _, p := range parts[1:] {
		if strings.HasPrefix(p, "E:") {
			val := strings.TrimPrefix(p, "E:")
			id, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return nil, ErrInvalidBallotQR
			}
			electionID = id
		} else if strings.HasPrefix(p, "C:") {
			val := strings.TrimPrefix(p, "C:")
			id, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return nil, ErrInvalidBallotQR
			}
			candidateID = id
		} else if strings.HasPrefix(p, "V:") {
			val := strings.TrimPrefix(p, "V:")
			v, err := strconv.Atoi(val)
			if err != nil {
				return nil, ErrInvalidBallotQR
			}
			version = v
		}
	}

	if electionID == 0 || candidateID == 0 || version == 0 {
		return nil, ErrInvalidBallotQR
	}

	return &BallotQR{
		ElectionID:  electionID,
		CandidateID: candidateID,
		Version:     version,
	}, nil
}
