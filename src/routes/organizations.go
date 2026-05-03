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
	orgs.Get("/:id/members", response.Wrap(handlers.ListOrgMembers))
	orgs.Put("/:id/members/:userId/role", response.Wrap(handlers.UpdateMemberRole))
	orgs.Get("/:id/roles", response.Wrap(handlers.ListOrgRoles))
	orgs.Post("/:id/roles", response.Wrap(handlers.CreateOrgRole))
	orgs.Delete("/:id/roles/:roleId", response.Wrap(handlers.DeleteOrgRole))
	orgs.Put("/:id/roles/:roleId/permissions", response.Wrap(handlers.UpdateOrgRolePermissions))
	orgs.Get("/:id/audit", response.Wrap(handlers.ListOrgAuditLog))
	orgs.Post("/", response.Wrap(handlers.CreateOrganization))
	orgs.Put("/:id", response.Wrap(handlers.UpdateOrganization))
	orgs.Patch("/:id/plan", response.Wrap(handlers.SetOrganizationPlan))
	orgs.Delete("/:id", response.Wrap(handlers.DeleteOrganization))
}
