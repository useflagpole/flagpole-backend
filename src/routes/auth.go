package routes

import (
	"flagpole/src/handlers"

	"github.com/gofiber/fiber/v3"
)

func registerAnonRoutes(api fiber.Router) {
	api.Post("/signup", handlers.Signup)
	api.Post("/login", handlers.Login)
}
