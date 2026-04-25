package routes

import (
	"flagpole/src/handlers"
	"flagpole/src/pkg/response"

	"github.com/gofiber/fiber/v3"
)

func registerOrganizationRoutes(api fiber.Router) {
	orgs := api.Group("/organizations")
	orgs.Get("/", response.Wrap(handlers.ListOrganizations))
	orgs.Get("/:id", response.Wrap(handlers.GetOrganization))
	orgs.Post("/", response.Wrap(handlers.CreateOrganization))
	orgs.Put("/:id", response.Wrap(handlers.UpdateOrganization))
	orgs.Patch("/:id/plan", response.Wrap(handlers.SetOrganizationPlan))
	orgs.Delete("/:id", response.Wrap(handlers.DeleteOrganization))
}
