package jwt

import (
	"fmt"
	"time"

	"maps"

	"github.com/golang-jwt/jwt/v5"
)

func NewToken(claims map[string]any, secret string, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	fmt.Printf("\nSECRET: %s, ", secret)
	fmt.Println(duration)

	token.Claims.(jwt.MapClaims)["iat"] = time.Now().Unix()
	token.Claims.(jwt.MapClaims)["exp"] = time.Now().Add(duration).Unix()
	token.Claims.(jwt.MapClaims)["iss"] = "calc-service"

	maps.Copy(token.Claims.(jwt.MapClaims), claims)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("can't signed token: %w", err)
	}

	return tokenString, nil
}
