package handlers

import (
	"errors"
	"strconv"

	"flagpole/src/controllers"
	"flagpole/src/dal"
	"flagpole/src/models"
	"flagpole/src/pkg/jwtutil"
	"flagpole/src/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type projectRequest struct {
	Name         string   `json:"name"`
	Environments []string `json:"environments"`
}

// ListProjects godoc
// @Summary      List projects for an organization
// @Tags         Projects
// @Produce      json
// @Param        org_id path int true "Organization ID"
// @Success      200 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Router       /organizations/{org_id}/projects [get]
func ListProjects(c fiber.Ctx) (int, response.APIResponse) {
	orgID, err := strconv.ParseUint(c.Params("org_id"), 10, 64)
	if err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid org id"}
	}

	if _, err := dal.Organization.GetByID(uint(orgID)); err != nil {
		return fiber.StatusNotFound, response.ErrorResponse{Error: "organization not found"}
	}

	if !dal.Organization.IsMember(uint(orgID), jwtutil.UserID(c)) {
		return fiber.StatusForbidden, response.ErrorResponse{Error: "forbidden"}
	}

	includeArchived := c.Query("getArchived") == "true"
	projects, err := dal.Project.ListByOrg(uint(orgID), includeArchived)
	if err != nil {
		return fiber.StatusInternalServerError, response.Error500
	}

	return fiber.StatusOK, response.DataResponse{Data: projects}
}

// CreateProject godoc
// @Summary      Create a project within an organization
// @Tags         Projects
// @Accept       json
// @Produce      json
// @Param        org_id path int true "Organization ID"
// @Param        body body projectRequest true "Project data"
// @Success      201 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Router       /organizations/{org_id}/projects [post]
func CreateProject(c fiber.Ctx) (int, response.APIResponse) {
	orgID, err := strconv.ParseUint(c.Params("org_id"), 10, 64)
	if err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid org id"}
	}

	if _, err := dal.Organization.GetByID(uint(orgID)); err != nil {
		return fiber.StatusNotFound, response.ErrorResponse{Error: "organization not found"}
	}

	if !dal.Organization.IsMember(uint(orgID), jwtutil.UserID(c)) {
		return fiber.StatusForbidden, response.ErrorResponse{Error: "forbidden"}
	}

	var req projectRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "couldn't parse body"}
	}
	if req.Name == "" {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "name is required"}
	}

	project, err := controllers.CreateProject(req.Name, uint(orgID), req.Environments)
	if err != nil {
		if errors.Is(err, controllers.ErrProjectLimitReached) {
			return fiber.StatusUnprocessableEntity, response.ErrorResponse{Error: err.Error()}
		}
		return fiber.StatusInternalServerError, response.Error500
	}

	logAudit(c, uint(orgID), &project.ID, models.ActionProjectCreate, project.Name, "Created project '"+project.Name+"'", "")
	return fiber.StatusCreated, response.DataResponse{Data: project}
}

// UpdateProject godoc
// @Summary      Rename a project
// @Tags         Projects
// @Accept       json
// @Produce      json
// @Param        org_id     path int true "Organization ID"
// @Param        project_id path int true "Project ID"
// @Param        body       body projectRequest true "Project data"
// @Success      200 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /organizations/{org_id}/projects/{project_id} [patch]
func UpdateProject(c fiber.Ctx) (int, response.APIResponse) {
	proj, status, errResp := resolveProject(c)
	if errResp != nil {
		return status, errResp
	}
	if status, errResp := requireAdmin(proj.OrganizationID, c); errResp != nil {
		return status, errResp
	}
	var req projectRequest
	if err := c.Bind().JSON(&req); err != nil || req.Name == "" {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "name is required"}
	}
	updated, err := controllers.RenameProject(proj, req.Name)
	if err != nil {
		return fiber.StatusInternalServerError, response.Error500
	}
	logAudit(c, proj.OrganizationID, &proj.ID, models.ActionProjectRename, updated.Name, "Renamed project to '"+updated.Name+"'", "")
	return fiber.StatusOK, response.DataResponse{Data: updated}
}

// ArchiveProject godoc
// @Summary      Archive a project
// @Tags         Projects
// @Produce      json
// @Param        org_id     path int true "Organization ID"
// @Param        project_id path int true "Project ID"
// @Success      200 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /organizations/{org_id}/projects/{project_id}/archive [post]
func ArchiveProject(c fiber.Ctx) (int, response.APIResponse) {
	proj, status, errResp := resolveProject(c)
	if errResp != nil {
		return status, errResp
	}
	if status, errResp := requireAdminOrEditor(proj.OrganizationID, c); errResp != nil {
		return status, errResp
	}
	updated, err := controllers.ArchiveProject(proj)
	if err != nil {
		return fiber.StatusInternalServerError, response.Error500
	}
	logAudit(c, proj.OrganizationID, &proj.ID, models.ActionProjectArchive, proj.Name, "Archived project '"+proj.Name+"'", "")
	return fiber.StatusOK, response.DataResponse{Data: updated}
}

// UnarchiveProject godoc
// @Summary      Unarchive a project
// @Tags         Projects
// @Produce      json
// @Param        org_id     path int true "Organization ID"
// @Param        project_id path int true "Project ID"
// @Success      200 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /organizations/{org_id}/projects/{project_id}/unarchive [post]
func UnarchiveProject(c fiber.Ctx) (int, response.APIResponse) {
	proj, status, errResp := resolveAnyProject(c)
	if errResp != nil {
		return status, errResp
	}
	if status, errResp := requireAdminOrEditor(proj.OrganizationID, c); errResp != nil {
		return status, errResp
	}
	updated, err := controllers.UnarchiveProject(proj)
	if err != nil {
		if errors.Is(err, controllers.ErrProjectLimitReached) {
			return fiber.StatusUnprocessableEntity, response.ErrorResponse{Error: err.Error()}
		}
		return fiber.StatusInternalServerError, response.Error500
	}
	logAudit(c, proj.OrganizationID, &proj.ID, models.ActionProjectUnarchive, proj.Name, "Unarchived project '"+proj.Name+"'", "")
	return fiber.StatusOK, response.DataResponse{Data: updated}
}
