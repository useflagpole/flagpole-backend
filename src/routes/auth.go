package routes

import (
	"time"

	"flagpole/src/handlers"
	"flagpole/src/pkg/response"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

var loginLimiter = limiter.New(limiter.Config{
	Max:        5,
	Expiration: 1 * time.Minute,
	KeyGenerator: func(c fiber.Ctx) string {
		return c.IP()
	},
	LimitReached: func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusTooManyRequests)
	},
})

func registerAnonRoutes(api fiber.Router) {
	api.Post("/signup", response.Wrap(handlers.Signup))
	api.Post("/login", loginLimiter, response.Wrap(handlers.Login))
}
