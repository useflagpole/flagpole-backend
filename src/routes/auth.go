package routes

import (
	"encoding/json"
	"time"

	"flagpole/src/config"
	"flagpole/src/controllers"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gofiber/fiber/v3"
)

type authPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func signup(c fiber.Ctx) error {
	body := new(authPayload)
	if err := json.Unmarshal(c.Body(), body); err != nil {
		return c.Status(400).SendString("couldn't parse body")
	}
	if len(body.Email) == 0 || len(body.Password) == 0 {
		return c.Status(400).SendString("email and password are required")
	}

	if err := controllers.RegisterUser(body.Email, body.Password); err != nil {
		return c.Status(409).SendString(err.Error())
	}

	return c.SendStatus(201)
}

func login(c fiber.Ctx) error {
	body := new(authPayload)
	if err := json.Unmarshal(c.Body(), body); err != nil {
		return c.Status(400).SendString("couldn't parse body")
	}
	if len(body.Email) == 0 || len(body.Password) == 0 {
		return c.Status(400).SendString("email and password are required")
	}

	if err := controllers.AuthenticateUser(body.Email, body.Password); err != nil {
		return c.Status(401).SendString(err.Error())
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": body.Email,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})

	signed, err := token.SignedString([]byte(config.Get().JWTSecret))
	if err != nil {
		return c.SendStatus(500)
	}

	return c.JSON(fiber.Map{"token": signed})
}
