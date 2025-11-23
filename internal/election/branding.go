package election

import (
	"errors"
	"strings"
	"time"
)

type BrandingSlot string

const (
	BrandingSlotPrimary   BrandingSlot = "primary"
	BrandingSlotSecondary BrandingSlot = "secondary"
)

var (
	ErrInvalidBrandingSlot  = errors.New("invalid branding slot")
	ErrBrandingFileNotFound = errors.New("branding file not found")
)

func ParseBrandingSlot(raw string) (BrandingSlot, error) {
	switch strings.ToLower(raw) {
	case string(BrandingSlotPrimary):
		return BrandingSlotPrimary, nil
	case string(BrandingSlotSecondary):
		return BrandingSlotSecondary, nil
	default:
		return "", ErrInvalidBrandingSlot
	}
}

type BrandingUser struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

type BrandingSettings struct {
	PrimaryLogoID   *string       `json:"primary_logo_id"`
	SecondaryLogoID *string       `json:"secondary_logo_id"`
	UpdatedAt       time.Time     `json:"updated_at"`
	UpdatedBy       *BrandingUser `json:"updated_by,omitempty"`
}

type BrandingFile struct {
	ID          string       `json:"id"`
	ElectionID  int64        `json:"-"`
	Slot        BrandingSlot `json:"slot"`
	ContentType string       `json:"content_type"`
	SizeBytes   int64        `json:"size"`
	Data        []byte       `json:"-"`
	CreatedAt   time.Time    `json:"created_at"`
	CreatedByID *int64       `json:"-"`
}

type BrandingFileCreate struct {
	ID          string
	ContentType string
	SizeBytes   int64
	Data        []byte
	CreatedByID int64
}
