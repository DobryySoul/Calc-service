package jwt

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewToken(t *testing.T) {
	secret := "test-secret"
	customClaims := map[string]any{
		"user_id": 123,
	}
	duration := 1 * time.Hour

	t.Run("successful token generation", func(t *testing.T) {
		tokenString, err := NewToken(customClaims, secret, duration)
		require.NoError(t, err)
		assert.NotEmpty(t, tokenString)

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		require.NoError(t, err)
		require.True(t, token.Valid)

		claims, ok := token.Claims.(jwt.MapClaims)
		require.True(t, ok)

		assert.Equal(t, ISSUER, claims["iss"])
		assert.NotZero(t, claims["iat"])
		assert.NotZero(t, claims["exp"])

		assert.Equal(t, float64(123), claims["user_id"])
	})

	t.Run("token with empty claims", func(t *testing.T) {
		tokenString, err := NewToken(map[string]any{}, secret, duration)
		require.NoError(t, err)
		assert.NotEmpty(t, tokenString)

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		require.NoError(t, err)
		assert.True(t, token.Valid)
	})

	t.Run("token with nil claims", func(t *testing.T) {
		tokenString, err := NewToken(nil, secret, duration)
		require.NoError(t, err)
		assert.NotEmpty(t, tokenString)

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		require.NoError(t, err)
		assert.True(t, token.Valid)
	})

	t.Run("token expiration", func(t *testing.T) {
		shortDuration := 1 * time.Millisecond
		tokenString, err := NewToken(customClaims, secret, shortDuration)
		require.NoError(t, err)

		time.Sleep(2 * time.Millisecond)

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		require.Error(t, err)
		assert.False(t, token.Valid)
	})

	t.Run("token validation with wrong secret", func(t *testing.T) {
		tokenString, err := NewToken(customClaims, secret, duration)
		require.NoError(t, err)

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte("wrong-secret"), nil
		})
		require.Error(t, err)
		assert.False(t, token.Valid)
	})

	t.Run("token with different signing method", func(t *testing.T) {
		token := jwt.New(jwt.SigningMethodRS256)
		claims := token.Claims.(jwt.MapClaims)
		claims["user_id"] = 123
		claims["iat"] = time.Now().Unix()
		claims["exp"] = time.Now().Add(duration).Unix()
		claims["iss"] = ISSUER

		_, err := token.SignedString([]byte(secret))
		require.Error(t, err)

		_, err = NewToken(customClaims, secret, duration)
		require.NoError(t, err)
	})

	t.Run("token with very long duration", func(t *testing.T) {
		longDuration := 10000 * time.Hour
		tokenString, err := NewToken(customClaims, secret, longDuration)
		require.NoError(t, err)

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		require.NoError(t, err)
		assert.True(t, token.Valid)

		claims := token.Claims.(jwt.MapClaims)
		exp := time.Unix(int64(claims["exp"].(float64)), 0)
		assert.True(t, exp.After(time.Now().Add(9999*time.Hour)))
	})
}

func TestNewTokenWithManyClaims(t *testing.T) {
	secret := "test-secret"
	duration := 1 * time.Hour

	claims := make(map[string]any)
	for i := 0; i < 1000; i++ {
		claims[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
	}

	tokenString, err := NewToken(claims, secret, duration)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	require.NoError(t, err)
	assert.True(t, token.Valid)
}

func TestNewTokenParallel(t *testing.T) {
	secret := "test-secret"
	duration := 1 * time.Hour

	t.Run("parallel token generation", func(t *testing.T) {
		t.Parallel()
		for i := 0; i < 10; i++ {
			i := i
			t.Run(fmt.Sprintf("token_%d", i), func(t *testing.T) {
				t.Parallel()
				claims := map[string]any{"id": i}
				token, err := NewToken(claims, secret, duration)
				require.NoError(t, err)
				assert.NotEmpty(t, token)
			})
		}
	})
}
