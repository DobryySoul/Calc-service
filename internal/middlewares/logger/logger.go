package logger

import (
	"log/slog"
	"net/http"
	"time"
)

func LoggerMiddleware(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		log.Info("request started",
			slog.String("method", r.Method),
			slog.String("url_path", r.URL.Path),
			slog.String("ip", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
			slog.String("host", r.Host),
		)

		next.ServeHTTP(w, r)

		log.Info("request completed",
			slog.Duration("duration", time.Since(start)))
	})
}
