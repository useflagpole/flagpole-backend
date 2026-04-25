package routes

import (
	"flagpole/src/handlers"
	"flagpole/src/pkg/response"

	"github.com/gofiber/fiber/v3"
)

func registerFlagRoutes(api fiber.Router) {
	api.Post("/flags/evaluate", response.Wrap(handlers.EvaluateFlags))
	api.Post("/flag", response.Wrap(handlers.AddFlag))
	api.Get("/flag/:flagname", response.Wrap(handlers.GetFlag))
	api.Put("/flag/:flagname", response.Wrap(handlers.SetFlag))
}
