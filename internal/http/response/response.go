package response

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type SuccessResponse struct {
	Data interface{} `json:"data"`
}

func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func Success(w http.ResponseWriter, statusCode int, data interface{}) {
	JSON(w, statusCode, SuccessResponse{Data: data})
}

func Error(w http.ResponseWriter, statusCode int, code, message string, details interface{}) {
	JSON(w, statusCode, ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	})
}

func BadRequest(w http.ResponseWriter, code, message string) {
	Error(w, http.StatusBadRequest, code, message, nil)
}

func Unauthorized(w http.ResponseWriter, code, message string) {
	Error(w, http.StatusUnauthorized, code, message, nil)
}

func Forbidden(w http.ResponseWriter, code, message string) {
	Error(w, http.StatusForbidden, code, message, nil)
}

func NotFound(w http.ResponseWriter, code, message string) {
	Error(w, http.StatusNotFound, code, message, nil)
}

func InternalServerError(w http.ResponseWriter, code, message string) {
	Error(w, http.StatusInternalServerError, code, message, nil)
}

func Conflict(w http.ResponseWriter, code, message string) {
	Error(w, http.StatusConflict, code, message, nil)
}

func UnprocessableEntity(w http.ResponseWriter, code, message string) {
	Error(w, http.StatusUnprocessableEntity, code, message, nil)
}
