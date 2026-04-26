package routes

import (
	"flagpole/src/handlers"
	"flagpole/src/pkg/response"

	"github.com/gofiber/fiber/v3"
)

func registerProjectRoutes(api fiber.Router) {
	api.Get("/organizations/:org_id/projects", response.Wrap(handlers.ListProjects))
	api.Post("/organizations/:org_id/projects", response.Wrap(handlers.CreateProject))
	api.Patch("/organizations/:org_id/projects/:project_id", response.Wrap(handlers.UpdateProject))
	api.Post("/organizations/:org_id/projects/:project_id/archive", response.Wrap(handlers.ArchiveProject))
	api.Post("/organizations/:org_id/projects/:project_id/unarchive", response.Wrap(handlers.UnarchiveProject))
	api.Get("/organizations/:org_id/projects/:project_id/audit", response.Wrap(handlers.ListProjectAuditLog))
}
