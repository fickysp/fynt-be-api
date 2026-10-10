package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	ID        string          `json:"id"`
	Role      string          `json:"role"`
	UpdatedAt jwt.NumericDate `json:"updated_at"`
	Type      string          `json:"type"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(id, role, secret string, updated_at time.Time) (string, error) {
	expired := 15 * time.Minute

	claims := &Claims{
		ID:        id,
		Role:      role,
		UpdatedAt: *jwt.NewNumericDate(updated_at),
		Type:      "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expired)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateRefreshToken(id, role, secret string, updated_at time.Time, rememberMe bool) (string, error) {
	expired := 24 * time.Hour
	if rememberMe {
		expired = 7 * 24 * time.Hour
	}
	claims := &Claims{
		ID:        id,
		Role:      role,
		UpdatedAt: *jwt.NewNumericDate(updated_at),
		Type:      "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expired)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
