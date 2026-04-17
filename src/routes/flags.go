package routes

import (
	"flagpole/src/handlers"

	"github.com/gofiber/fiber/v3"
)

func registerFlagRoutes(api fiber.Router) {
	api.Post("/flags/evaluate", handlers.EvaluateFlags)
	api.Post("/flag", handlers.AddFlag)
	api.Get("/flag/:flagname", handlers.GetFlag)
	api.Put("/flag/:flagname", handlers.SetFlag)
}
