package jwtutil

import (
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func Claims(c fiber.Ctx) jwt.MapClaims {
	claims, _ := c.Locals("claims").(jwt.MapClaims)
	return claims
}

func UserID(c fiber.Ctx) uuid.UUID {
	id, _ := c.Locals("userID").(uuid.UUID)
	return id
}
