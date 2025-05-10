package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func LoggerMiddleware(log *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			method := r.Method
			path := r.URL.Path
			addr := r.RemoteAddr

			next.ServeHTTP(w, r)

			duration := time.Since(start)

			log.Info("request completed",
				zap.String("method", method),
				zap.String("path", path),
				zap.String("addr", addr),
				zap.Duration("duration", duration))
		})
	}
}
