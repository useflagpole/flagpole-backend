package routes

import (
	"flagpole/src/handlers"
	"flagpole/src/pkg/response"

	"github.com/gofiber/fiber/v3"
)

func registerUserRoutes(api fiber.Router) {
	users := api.Group("/users")
	users.Get("/:user_id", response.Wrap(handlers.GetUser))
	users.Get("/:user_id/organizations", response.Wrap(handlers.ListUserOrganizations))
	users.Patch("/:user_id/username", response.Wrap(handlers.UpdateUsername))
}
