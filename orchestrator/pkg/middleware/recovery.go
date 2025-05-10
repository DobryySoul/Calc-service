package middleware

import (
	"net/http"

	"go.uber.org/zap"
)

func RecoveryMiddleware(log *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Error("recovery error", zap.Any("err", err))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}