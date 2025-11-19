package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"pemira-api/internal/shared/ctxkeys"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &responseWriter{
			ResponseWriter: w,
			status:         200,
		}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		fields := []interface{}{
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.status,
			"duration_ms", duration.Milliseconds(),
			"size", wrapped.size,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		}

		// Add user info if available
		if userID, ok := r.Context().Value(ctxkeys.UserIDKey).(int64); ok {
			fields = append(fields, "user_id", userID)
		}
		if role, ok := r.Context().Value(ctxkeys.UserRoleKey).(string); ok {
			fields = append(fields, "role", role)
		}

		slog.Info("http_request", fields...)
	})
}
