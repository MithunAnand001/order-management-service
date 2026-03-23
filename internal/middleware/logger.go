package middleware

import (
	"net/http"
	"time"

	"order-management-service/internal/utils"

	"go.uber.org/zap"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func LoggerMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := utils.Now()
			reqID := utils.GetRequestID(r.Context())

			rw := &responseWriter{w, http.StatusOK}
			next.ServeHTTP(rw, r)

			duration := time.Since(start)

			logger.Info("HTTP Request",
				zap.String("request_id", reqID),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", rw.status),
				zap.Duration("latency", duration),
			)
		})
	}
}
