package service

import (
	"context"
	"fmt"
	"time"

	"github.com/DobryySoul/Calc-service/internal/http/models"
	"github.com/DobryySoul/Calc-service/pkg/jwt"
	"github.com/DobryySoul/Calc-service/pkg/utils"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthRepo interface {
	Register(ctx context.Context, user *models.User) error
	Login(ctx context.Context, email, password string) (*models.User, error)
}

type AuthService struct {
	authRepo    AuthRepo
	tokenSecret string
	tokenTTL    time.Duration
	log         *zap.Logger
}

func NewAuthService(repo AuthRepo, tokenSecret string, tokenTTL string, log *zap.Logger) *AuthService {
	TTL, _ := time.ParseDuration(tokenTTL + "s")
	return &AuthService{
		authRepo:    repo,
		log:         log,
		tokenSecret: tokenSecret,
		tokenTTL:    TTL,
	}
}

func (a *AuthService) Register(ctx context.Context, email, password string) error {
	a.log.Info("Start registration user")

	user := &models.User{
		Email:    email,
		Password: password,
	}

	if err := utils.ValidateUserCredentials(user); err != nil {
		return fmt.Errorf("failed to validate user: %w", err)
	}

	hashPass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to generate password hash: %w", err)
	}

	user.Password = string(hashPass)
	if err = a.authRepo.Register(ctx, user); err != nil {
		return fmt.Errorf("failed to register user: %w", err)
	}

	a.log.Info("User registered successfully")
	return nil
}

func (a *AuthService) Login(ctx context.Context, email, password string) (*models.User, string, error) {
	a.log.Info("Start login user")

	user, err := a.authRepo.Login(ctx, email, password)
	if err != nil {
		return nil, "", fmt.Errorf("failed to login user: %w", err)
	}

	token, err := a.doToken(user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", fmt.Errorf("failed to compare password: %w", err)
	}

	a.log.Info("User logged in successfully")
	return user, token, nil
}

func (a *AuthService) Logout(ctx context.Context, user *models.User) error {
	a.log.Info("Start logout user")
	return nil
}

func (a *AuthService) doToken(userId uint64) (string, error) {
	payload := map[string]any{
		"uid": userId,
	}

	token, err := jwt.NewToken(payload, a.tokenSecret, a.tokenTTL)
	if err != nil {
		return "", fmt.Errorf("can't generate token: %w", err)
	}

	return token, nil
}
