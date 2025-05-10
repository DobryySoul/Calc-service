package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/DobryySoul/orchestrator/internal/http/models/resp"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)

type userClaim struct {
	jwt.RegisteredClaims
	Uid uint64 `json:"uid"`
}

func AuthMiddleware(secret string, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v1/login" || r.URL.Path == "/api/v1/register" {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("Content-Type", "application/json")

			var responseError resp.ResponseError

			tokenString := r.Header.Get("Authorization")
			if tokenString == "" {
				logger.Error("no token provided")
				sendErrorResponse(w, http.StatusUnauthorized, ErrUnauthorized, &responseError)
				return
			}

			tokenString = strings.TrimPrefix(tokenString, "Bearer ")
			token, err := jwt.ParseWithClaims(tokenString, &userClaim{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})

			if err != nil {
				if errors.Is(err, jwt.ErrTokenExpired) {
					logger.Error("token expired")
					sendErrorResponse(w, http.StatusUnauthorized, ErrExpiredToken, &responseError)
					return
				}
				logger.Error("invalid token", zap.Error(err))
				sendErrorResponse(w, http.StatusUnauthorized, ErrInvalidToken, &responseError)
				return
			}

			if !token.Valid {
				logger.Error("invalid token")
				sendErrorResponse(w, http.StatusUnauthorized, ErrInvalidToken, &responseError)
				return
			}

			if claims, ok := token.Claims.(*userClaim); ok {
				ctx := context.WithValue(r.Context(), "uid", claims.Uid)
				r = r.WithContext(ctx)

				uid := strconv.FormatUint(claims.Uid, 10)
				http.SetCookie(w, &http.Cookie{
					Name:     "user_id",
					Value:    uid,
					Path:     "/",
					HttpOnly: true,
					Secure:   false,
					SameSite: http.SameSiteStrictMode,
				})
			}

			next.ServeHTTP(w, r)
		})
	}
}

func sendErrorResponse(w http.ResponseWriter, status int, err error, responseError *resp.ResponseError) {
	w.WriteHeader(status)
	responseError.Error = err.Error()
	_ = json.NewEncoder(w).Encode(responseError)
}
