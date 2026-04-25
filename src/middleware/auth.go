package middleware

import (
	"strings"

	"flagpole/src/config"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func Auth(c fiber.Ctx) error {
	header := c.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return c.Status(401).SendString("missing or invalid authorization header")
	}

	tokenStr := strings.TrimPrefix(header, "Bearer ")

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return []byte(config.Get().JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return c.Status(401).SendString("invalid or expired token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(401).SendString("invalid token claims")
	}

	userIDStr, _ := claims["userId"].(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(401).SendString("invalid token: bad user id")
	}

	c.Locals("claims", claims)
	c.Locals("userID", userID)
	return c.Next()
}
