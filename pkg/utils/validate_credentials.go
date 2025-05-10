package utils

import (
	"fmt"
	"strings"

	"github.com/DobryySoul/Calc-service/internal/http/models"
)

func ValidateUserCredentials(user *models.User) error {
	if !strings.Contains(user.Email, "@") {
		return fmt.Errorf("invalid email")
	}

	if err := validatePassword(user.Password); err != nil {
		return fmt.Errorf("invalid password: %v", err)
	}

	return nil
}

func validatePassword(pass string) any {

	if len(pass) <= 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	if !strings.ContainsAny(pass, "!@#$%^&*()-_=+[]{}|;:,.<>?") {
		return fmt.Errorf("password must contain at least one special character")
	}

	if !strings.ContainsAny(pass, "0123456789") {
		return fmt.Errorf("password must contain at least one number")
	}

	if !strings.ContainsAny(pass, "abcdefghijklmnopqrstuvwxyz") {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if !strings.ContainsAny(pass, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	return nil
}
