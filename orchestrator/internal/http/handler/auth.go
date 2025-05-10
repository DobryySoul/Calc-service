package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/DobryySoul/orchestrator/internal/service"
	"go.uber.org/zap"
)

type authHandler struct {
	auth *service.AuthService
	log  *zap.Logger
}

func NewAuthHandler(log *zap.Logger, authService *service.AuthService) *authHandler {
	return &authHandler{
		auth: authService,
		log:  log,
	}
}

type doLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type doLoginResponse struct {
	Token  string `json:"token"`
	UserID uint64 `json:"userId"`
}

func (a *authHandler) Login(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var req doLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.log.Error("failed to decode request", zap.Error(err))
		http.Error(w, "failed to decode request", http.StatusBadRequest)

		return
	}

	user, token, err := a.auth.Login(r.Context(), req.Email, req.Password)
	if errors.Is(err, ErrUserNotFound) {
		http.Error(w, "user not found", http.StatusConflict)

		return
	}
	if errors.Is(err, ErrInvalidCredentials) {
		http.Error(w, "wrong credentials", http.StatusUnauthorized)

		return
	}
	if err != nil {
		a.log.Error("failed to login user", zap.Error(err))
		http.Error(w, "auth service problems", http.StatusInternalServerError)

		return
	}

	resp := doLoginResponse{
		Token:  token,
		UserID: uint64(user.ID),
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		a.log.Error("failed to encode response", zap.Error(err))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
}

type doRegisterNewUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (a *authHandler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req doRegisterNewUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.log.Error("failed to decode request", zap.Error(err))
		http.Error(w, "failed to decode request", http.StatusBadRequest)

		return
	}

	if err := a.auth.Register(r.Context(), req.Email, req.Password); err != nil {
		a.log.Error("failed to register user", zap.Error(err))
		http.Error(w, "failed to register user", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
}
