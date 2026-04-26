package handlers

import (
	"log"
	"strconv"

	"flagpole/src/dal"
	"flagpole/src/pkg/jwtutil"
	"flagpole/src/pkg/response"

	"github.com/gofiber/fiber/v3"
)

func logAudit(c fiber.Ctx, orgID uint, projectID *uint, action, target, detail, env string) {
	if err := dal.Audit.LogWithActor(orgID, projectID, jwtutil.UserID(c), action, target, detail, env); err != nil {
		log.Printf("logAudit: %v", err)
	}
}

// ListProjectAuditLog godoc
// @Summary      List audit log entries for a project
// @Tags         Audit
// @Produce      json
// @Param        org_id     path int true "Organization ID"
// @Param        project_id path int true "Project ID"
// @Success      200 {object} response.DataResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Router       /organizations/{org_id}/projects/{project_id}/audit [get]
func ListProjectAuditLog(c fiber.Ctx) (int, response.APIResponse) {
	proj, status, errResp := resolveProject(c)
	if errResp != nil {
		return status, errResp
	}
	entries, err := dal.Audit.ListByProject(proj.ID)
	if err != nil {
		log.Printf("ListProjectAuditLog: %v", err)
		return fiber.StatusInternalServerError, response.Error500
	}
	return fiber.StatusOK, response.DataResponse{Data: entries}
}

// ListOrgAuditLog godoc
// @Summary      List audit log entries for an organization
// @Tags         Audit
// @Produce      json
// @Param        id path int true "Organization ID"
// @Success      200 {object} response.DataResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Router       /organizations/{id}/audit [get]
func ListOrgAuditLog(c fiber.Ctx) (int, response.APIResponse) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid id"}
	}
	if _, err := dal.Organization.GetByID(uint(id)); err != nil {
		return fiber.StatusNotFound, response.ErrorResponse{Error: "organization not found"}
	}
	if !dal.Organization.IsMember(uint(id), jwtutil.UserID(c)) {
		return fiber.StatusForbidden, response.ErrorResponse{Error: "forbidden"}
	}
	entries, err := dal.Audit.ListByOrg(uint(id))
	if err != nil {
		return fiber.StatusInternalServerError, response.Error500
	}
	return fiber.StatusOK, response.DataResponse{Data: entries}
}
