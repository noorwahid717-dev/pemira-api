package tps

import (
	"errors"
	"net/http"
)

var (
	ErrTPSNotFound          = errors.New("TPS tidak ditemukan")
	ErrTPSInactive          = errors.New("TPS belum/tidak aktif")
	ErrTPSClosed            = errors.New("TPS sudah ditutup")
	ErrQRInvalid            = errors.New("Payload QR tidak valid")
	ErrQRRevoked            = errors.New("QR sudah tidak berlaku")
	ErrElectionNotOpen      = errors.New("Pemilu bukan di fase voting")
	ErrNotEligible          = errors.New("Mahasiswa bukan DPT / tidak berhak")
	ErrAlreadyVoted         = errors.New("Mahasiswa sudah pernah voting")
	ErrCheckinNotFound      = errors.New("Data check-in tidak ada")
	ErrCheckinNotPending    = errors.New("Check-in bukan status PENDING")
	ErrCheckinExpired       = errors.New("Check-in sudah kadaluarsa")
	ErrCheckinAlreadyExists = errors.New("Check-in sudah ada")
	ErrTPSAccessDenied      = errors.New("Panitia TPS tidak di-assign ke TPS ini")
	ErrTPSCodeDuplicate     = errors.New("Kode TPS sudah digunakan")
	ErrInvalidTimeFormat    = errors.New("Format waktu tidak valid")
	ErrNotTPSVoter          = errors.New("Pemilih bukan TPS")
	ErrTPSMismatch          = errors.New("TPS tidak sesuai")
	ErrOperatorExists       = errors.New("Operator sudah ada")
	ErrOperatorNotFound     = errors.New("Operator tidak ditemukan")
)

type ErrorCode struct {
	Code       string
	HTTPStatus int
}

var errorCodeMap = map[error]ErrorCode{
	ErrTPSNotFound:          {Code: "TPS_NOT_FOUND", HTTPStatus: http.StatusNotFound},
	ErrTPSInactive:          {Code: "TPS_INACTIVE", HTTPStatus: http.StatusBadRequest},
	ErrTPSClosed:            {Code: "TPS_CLOSED", HTTPStatus: http.StatusBadRequest},
	ErrQRInvalid:            {Code: "QR_INVALID", HTTPStatus: http.StatusBadRequest},
	ErrQRRevoked:            {Code: "QR_REVOKED", HTTPStatus: http.StatusBadRequest},
	ErrElectionNotOpen:      {Code: "ELECTION_NOT_OPEN", HTTPStatus: http.StatusBadRequest},
	ErrNotEligible:          {Code: "NOT_ELIGIBLE", HTTPStatus: http.StatusBadRequest},
	ErrAlreadyVoted:         {Code: "ALREADY_VOTED", HTTPStatus: http.StatusConflict},
	ErrCheckinNotFound:      {Code: "CHECKIN_NOT_FOUND", HTTPStatus: http.StatusNotFound},
	ErrCheckinNotPending:    {Code: "CHECKIN_NOT_PENDING", HTTPStatus: http.StatusBadRequest},
	ErrCheckinExpired:       {Code: "CHECKIN_EXPIRED", HTTPStatus: http.StatusBadRequest},
	ErrCheckinAlreadyExists: {Code: "CHECKIN_EXISTS", HTTPStatus: http.StatusBadRequest},
	ErrTPSAccessDenied:      {Code: "TPS_ACCESS_DENIED", HTTPStatus: http.StatusForbidden},
	ErrTPSCodeDuplicate:     {Code: "TPS_CODE_DUPLICATE", HTTPStatus: http.StatusConflict},
	ErrInvalidTimeFormat:    {Code: "INVALID_TIME_FORMAT", HTTPStatus: http.StatusBadRequest},
	ErrNotTPSVoter:          {Code: "NOT_TPS_VOTER", HTTPStatus: http.StatusBadRequest},
	ErrTPSMismatch:          {Code: "TPS_MISMATCH", HTTPStatus: http.StatusBadRequest},
	ErrOperatorExists:       {Code: "OPERATOR_EXISTS", HTTPStatus: http.StatusConflict},
	ErrOperatorNotFound:     {Code: "OPERATOR_NOT_FOUND", HTTPStatus: http.StatusNotFound},
}

func GetErrorCode(err error) (string, int) {
	if ec, ok := errorCodeMap[err]; ok {
		return ec.Code, ec.HTTPStatus
	}
	return "INTERNAL_ERROR", http.StatusInternalServerError
}
