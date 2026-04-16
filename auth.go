package main

import (
	"encoding/json"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gofiber/fiber/v3"
)

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func SignupRoute(c fiber.Ctx) error {
	body := new(AuthPayload)
	if err := json.Unmarshal(c.Body(), body); err != nil {
		return c.Status(400).SendString("couldn't parse body")
	}
	if len(body.Email) == 0 || len(body.Password) == 0 {
		return c.Status(400).SendString("email and password are required")
	}

	if err := UserStore.Register(body.Email, body.Password); err != nil {
		return c.Status(409).SendString(err.Error())
	}

	return c.SendStatus(201)
}

func LoginRoute(c fiber.Ctx) error {
	body := new(AuthPayload)
	if err := json.Unmarshal(c.Body(), body); err != nil {
		return c.Status(400).SendString("couldn't parse body")
	}
	if len(body.Email) == 0 || len(body.Password) == 0 {
		return c.Status(400).SendString("email and password are required")
	}

	if err := UserStore.Authenticate(body.Email, body.Password); err != nil {
		return c.Status(401).SendString(err.Error())
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": body.Email,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})

	signed, err := token.SignedString([]byte(flagReader.jwtSecret))
	if err != nil {
		return c.SendStatus(500)
	}

	return c.JSON(fiber.Map{"token": signed})
}
