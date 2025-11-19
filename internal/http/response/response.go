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

func BadRequest(w http.ResponseWriter, message string, details interface{}) {
	Error(w, http.StatusBadRequest, "BAD_REQUEST", message, details)
}

func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, "UNAUTHORIZED", message, nil)
}

func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, "FORBIDDEN", message, nil)
}

func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, "NOT_FOUND", message, nil)
}

func InternalServerError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", message, nil)
}
