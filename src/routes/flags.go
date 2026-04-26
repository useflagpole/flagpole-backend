package routes

import (
	"flagpole/src/handlers"
	"flagpole/src/pkg/response"

	"github.com/gofiber/fiber/v3"
)

func registerFlagRoutes(api fiber.Router) {
	proj := "/organizations/:org_id/projects/:project_id"
	api.Get(proj+"/flags", response.Wrap(handlers.ListFlags))
	api.Post(proj+"/flags", response.Wrap(handlers.CreateFlag))
	api.Get(proj+"/flags/:flag_id", response.Wrap(handlers.GetFlag))
	api.Patch(proj+"/flags/:flag_id", response.Wrap(handlers.UpdateFlag))
	api.Delete(proj+"/flags/:flag_id", response.Wrap(handlers.DeleteFlag))
}
