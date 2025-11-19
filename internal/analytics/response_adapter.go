package analytics

import (
"encoding/json"
"net/http"
)

// StandardResponseWriter adapts the internal/http/response package
type StandardResponseWriter struct{}

// NewStandardResponseWriter creates a new standard response writer adapter
func NewStandardResponseWriter() ResponseWriter {
return &StandardResponseWriter{}
}

// Success sends a success response with data
func (s *StandardResponseWriter) Success(w http.ResponseWriter, statusCode int, data interface{}) {
type SuccessResponse struct {
Data interface{} `json:"data"`
}

w.Header().Set("Content-Type", "application/json")
w.WriteHeader(statusCode)
json.NewEncoder(w).Encode(SuccessResponse{Data: data})
}

// BadRequest sends a bad request error response
func (s *StandardResponseWriter) BadRequest(w http.ResponseWriter, message string, details interface{}) {
type ErrorResponse struct {
Code    string      `json:"code"`
Message string      `json:"message"`
Details interface{} `json:"details,omitempty"`
}

w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusBadRequest)
json.NewEncoder(w).Encode(ErrorResponse{
Code:    "VALIDATION_ERROR",
Message: message,
Details: details,
})
}

// InternalServerError sends an internal server error response
func (s *StandardResponseWriter) InternalServerError(w http.ResponseWriter, message string) {
type ErrorResponse struct {
Code    string `json:"code"`
Message string `json:"message"`
}

w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusInternalServerError)
json.NewEncoder(w).Encode(ErrorResponse{
Code:    "INTERNAL_ERROR",
Message: message,
})
}

// NotFound sends a not found error response
func (s *StandardResponseWriter) NotFound(w http.ResponseWriter, message string) {
type ErrorResponse struct {
Code    string `json:"code"`
Message string `json:"message"`
}

w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusNotFound)
json.NewEncoder(w).Encode(ErrorResponse{
Code:    "NOT_FOUND",
Message: message,
})
}
