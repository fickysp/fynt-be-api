package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	ID        string          `json:"id"`
	Role      string          `json:"role"`
	ProfileID string          `json:"profile_id"`
	UpdatedAt jwt.NumericDate `json:"updated_at"`
	Type      string          `json:"type"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(id, role, profID, secret string, updated_at time.Time) (string, error) {
	expired := 15 * time.Minute

	claims := &Claims{
		ID:        id,
		Role:      role,
		ProfileID: profID,
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

func GenerateRefreshToken(id, role, profID, secret string, updated_at time.Time, rememberMe bool) (string, error) {
	expired := 24 * time.Hour
	if rememberMe {
		expired = 7 * 24 * time.Hour
	}
	claims := &Claims{
		ID:        id,
		Role:      role,
		ProfileID: profID,
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

func ValidateToken(tokens, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokens, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("signing method invalid")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, errors.New("Token invalid")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("Token invalid")
	}

	return claims, nil
}
