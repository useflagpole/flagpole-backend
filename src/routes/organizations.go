package routes

import (
	"flagpole/src/handlers"

	"github.com/gofiber/fiber/v3"
)

func registerOrganizationRoutes(api fiber.Router) {
	orgs := api.Group("/organizations")
	orgs.Get("/", handlers.ListOrganizations)
	orgs.Get("/:id", handlers.GetOrganization)
	orgs.Post("/", handlers.CreateOrganization)
	orgs.Put("/:id", handlers.UpdateOrganization)
	orgs.Delete("/:id", handlers.DeleteOrganization)
}
