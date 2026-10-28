package middleware

import (
	"strings"

	"github.com/fickysp/fynt-be-api/config"
	myjwt "github.com/fickysp/fynt-be-api/pkg/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

type AuthMiddleware interface {
	RequireToken() fiber.Handler
	RequireRole(allowedRoles ...string) fiber.Handler
}

type authMiddleware struct {
	cfg *config.Config
}

func NewAuthMiddleware(cfg *config.Config) AuthMiddleware {
	return &authMiddleware{
		cfg: cfg,
	}
}

func (m *authMiddleware) RequireToken() fiber.Handler {
	return func(c fiber.Ctx) error {
		var tokenString string

		authHeader := c.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			tokenString = c.Cookies("access_token")
		}

		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Missing authorization token",
			})
		}

		token, err := jwt.ParseWithClaims(tokenString, &myjwt.Claims{}, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return []byte(m.cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Token invalid or expired",
			})
		}

		claims, ok := token.Claims.(*myjwt.Claims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Failed to parse token claims",
			})
		}

		c.Locals("account_id", claims.ID)
		c.Locals("role", claims.Role)
		c.Locals("profile_id", claims.ProfileID)

		return c.Next()
	}
}

func (m *authMiddleware) RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		userRole, ok := c.Locals("role").(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "User role not found in session",
			})
		}

		for _, role := range allowedRoles {
			if role == userRole {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Access denied for this role",
		})
	}
}
