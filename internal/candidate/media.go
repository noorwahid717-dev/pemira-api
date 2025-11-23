package candidate

import (
	"errors"
	"strings"
	"time"
)

type CandidateMediaSlot string

const (
	CandidateMediaSlotProfile    CandidateMediaSlot = "profile"
	CandidateMediaSlotPoster     CandidateMediaSlot = "poster"
	CandidateMediaSlotPhotoExtra CandidateMediaSlot = "photo_extra"
	CandidateMediaSlotPDFProgram CandidateMediaSlot = "pdf_program"
	CandidateMediaSlotPDFVisimisi CandidateMediaSlot = "pdf_visimisi"
)

var (
	ErrInvalidCandidateMediaSlot = errors.New("invalid candidate media slot")
	ErrCandidateMediaNotFound    = errors.New("candidate media not found")
)

func ParseCandidateMediaSlot(raw string) (CandidateMediaSlot, error) {
	switch strings.ToLower(raw) {
	case string(CandidateMediaSlotProfile):
		return CandidateMediaSlotProfile, nil
	case string(CandidateMediaSlotPoster):
		return CandidateMediaSlotPoster, nil
	case string(CandidateMediaSlotPhotoExtra):
		return CandidateMediaSlotPhotoExtra, nil
	case string(CandidateMediaSlotPDFProgram):
		return CandidateMediaSlotPDFProgram, nil
	case string(CandidateMediaSlotPDFVisimisi):
		return CandidateMediaSlotPDFVisimisi, nil
	default:
		return "", ErrInvalidCandidateMediaSlot
	}
}

type CandidateMedia struct {
	ID            string              `json:"id"`
	CandidateID   int64               `json:"candidate_id"`
	Slot          CandidateMediaSlot  `json:"slot"`
	FileName      string              `json:"file_name"`
	ContentType   string              `json:"content_type"`
	SizeBytes     int64               `json:"size"`
	Data          []byte              `json:"-"`
	CreatedAt     time.Time           `json:"created_at"`
	CreatedByID   *int64              `json:"created_by_admin_id,omitempty"`
}

type CandidateMediaMeta struct {
	ID          string             `json:"id"`
	Slot        CandidateMediaSlot `json:"slot"`
	Label       string             `json:"label"`
	ContentType string             `json:"content_type"`
	SizeBytes   int64              `json:"size"`
	CreatedAt   time.Time          `json:"created_at"`
}

type CandidateMediaCreate struct {
	ID          string
	Slot        CandidateMediaSlot
	FileName    string
	ContentType string
	SizeBytes   int64
	Data        []byte
	CreatedByID int64
}
