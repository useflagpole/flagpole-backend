package routes

import (
	"flagpole/src/handlers"
	"flagpole/src/pkg/response"

	"github.com/gofiber/fiber/v3"
)

func registerEnvironmentRoutes(api fiber.Router) {
	proj := "/organizations/:org_id/projects/:project_id"
	api.Get(proj+"/environments", response.Wrap(handlers.ListEnvironments))
	api.Post(proj+"/environments", response.Wrap(handlers.CreateEnvironment))
	api.Patch(proj+"/environments/:env_id", response.Wrap(handlers.RenameEnvironment))
	api.Delete(proj+"/environments/:env_id", response.Wrap(handlers.DeleteEnvironment))
}
