package handlers

import (
	"errors"
	"strconv"

	"flagpole/src/controllers"
	"flagpole/src/dal"
	"flagpole/src/models"
	"flagpole/src/pkg/jwtutil"
	"flagpole/src/pkg/permissions"
	"flagpole/src/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type envNameRequest struct {
	Name string `json:"name"`
}

func requirePermission(orgID uint, perm string, c fiber.Ctx) (int, response.APIResponse) {
	if !dal.Organization.HasPermission(orgID, jwtutil.UserID(c), perm) {
		return fiber.StatusForbidden, response.ErrorResponse{Error: "forbidden"}
	}
	return 0, nil
}

func resolveProject(c fiber.Ctx) (*models.Project, int, response.APIResponse) {
	orgID, err := strconv.ParseUint(c.Params("org_id"), 10, 64)
	if err != nil {
		return nil, fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid org id"}
	}
	if _, err := dal.Organization.GetByID(uint(orgID)); err != nil {
		return nil, fiber.StatusNotFound, response.ErrorResponse{Error: "organization not found"}
	}
	if !dal.Organization.IsMember(uint(orgID), jwtutil.UserID(c)) {
		return nil, fiber.StatusForbidden, response.ErrorResponse{Error: "forbidden"}
	}
	projectID, err := strconv.ParseUint(c.Params("project_id"), 10, 64)
	if err != nil {
		return nil, fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid project id"}
	}
	proj, err := dal.Project.GetByID(uint(projectID))
	if err != nil {
		return nil, fiber.StatusNotFound, response.ErrorResponse{Error: "project not found"}
	}
	if proj.OrganizationID != uint(orgID) {
		return nil, fiber.StatusForbidden, response.ErrorResponse{Error: "forbidden"}
	}
	if !proj.IsActive {
		return nil, fiber.StatusNotFound, response.ErrorResponse{Error: "project not found"}
	}
	return proj, 0, nil
}

// resolveAnyProject resolves a project regardless of its archived state.
// Use only for the unarchive endpoint.
func resolveAnyProject(c fiber.Ctx) (*models.Project, int, response.APIResponse) {
	orgID, err := strconv.ParseUint(c.Params("org_id"), 10, 64)
	if err != nil {
		return nil, fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid org id"}
	}
	if _, err := dal.Organization.GetByID(uint(orgID)); err != nil {
		return nil, fiber.StatusNotFound, response.ErrorResponse{Error: "organization not found"}
	}
	if !dal.Organization.IsMember(uint(orgID), jwtutil.UserID(c)) {
		return nil, fiber.StatusForbidden, response.ErrorResponse{Error: "forbidden"}
	}
	projectID, err := strconv.ParseUint(c.Params("project_id"), 10, 64)
	if err != nil {
		return nil, fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid project id"}
	}
	proj, err := dal.Project.GetByID(uint(projectID))
	if err != nil {
		return nil, fiber.StatusNotFound, response.ErrorResponse{Error: "project not found"}
	}
	if proj.OrganizationID != uint(orgID) {
		return nil, fiber.StatusForbidden, response.ErrorResponse{Error: "forbidden"}
	}
	return proj, 0, nil
}

func envErr(err error) (int, response.APIResponse) {
	switch {
	case errors.Is(err, controllers.ErrMaxEnvironments):
		return fiber.StatusUnprocessableEntity, response.ErrorResponse{Error: err.Error()}
	case errors.Is(err, controllers.ErrEnvAlreadyExists):
		return fiber.StatusConflict, response.ErrorResponse{Error: err.Error()}
	case errors.Is(err, controllers.ErrEnvNotFound):
		return fiber.StatusNotFound, response.ErrorResponse{Error: err.Error()}
	case errors.Is(err, controllers.ErrEnvProtected):
		return fiber.StatusForbidden, response.ErrorResponse{Error: err.Error()}
	default:
		return fiber.StatusInternalServerError, response.Error500
	}
}

// ListEnvironments godoc
// @Summary      List environments for a project
// @Tags         Environments
// @Produce      json
// @Param        org_id     path int true "Organization ID"
// @Param        project_id path int true "Project ID"
// @Success      200 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Router       /organizations/{org_id}/projects/{project_id}/environments [get]
func ListEnvironments(c fiber.Ctx) (int, response.APIResponse) {
	proj, status, errResp := resolveProject(c)
	if errResp != nil {
		return status, errResp
	}
	envs, err := controllers.ListEnvironments(proj)
	if err != nil {
		return fiber.StatusInternalServerError, response.Error500
	}
	return fiber.StatusOK, response.DataResponse{Data: envs}
}

// CreateEnvironment godoc
// @Summary      Create an environment in a project
// @Tags         Environments
// @Accept       json
// @Produce      json
// @Param        org_id     path int true "Organization ID"
// @Param        project_id path int true "Project ID"
// @Param        body       body envNameRequest true "Environment name"
// @Success      201 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      409 {object} response.ErrorResponse
// @Failure      422 {object} response.ErrorResponse
// @Router       /organizations/{org_id}/projects/{project_id}/environments [post]
func CreateEnvironment(c fiber.Ctx) (int, response.APIResponse) {
	proj, status, errResp := resolveProject(c)
	if errResp != nil {
		return status, errResp
	}
	if status, errResp := requirePermission(proj.OrganizationID, permissions.EnvCreate, c); errResp != nil {
		return status, errResp
	}
	var req envNameRequest
	if err := c.Bind().JSON(&req); err != nil || req.Name == "" {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "name is required"}
	}
	envs, err := controllers.CreateEnvironment(proj, req.Name)
	if err != nil {
		return envErr(err)
	}
	return fiber.StatusCreated, response.DataResponse{Data: envs}
}

// RenameEnvironment godoc
// @Summary      Rename an environment
// @Tags         Environments
// @Accept       json
// @Produce      json
// @Param        org_id     path int    true "Organization ID"
// @Param        project_id path int    true "Project ID"
// @Param        env_name   path string true "Current environment name"
// @Param        body       body envNameRequest true "New name"
// @Success      200 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Router       /organizations/{org_id}/projects/{project_id}/environments/{env_name} [patch]
func RenameEnvironment(c fiber.Ctx) (int, response.APIResponse) {
	proj, status, errResp := resolveProject(c)
	if errResp != nil {
		return status, errResp
	}
	if status, errResp := requirePermission(proj.OrganizationID, permissions.EnvRename, c); errResp != nil {
		return status, errResp
	}
	var req envNameRequest
	if err := c.Bind().JSON(&req); err != nil || req.Name == "" {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "name is required"}
	}
	envs, err := controllers.RenameEnvironment(proj, c.Params("env_name"), req.Name)
	if err != nil {
		return envErr(err)
	}
	return fiber.StatusOK, response.DataResponse{Data: envs}
}

// DeleteEnvironment godoc
// @Summary      Delete an environment
// @Tags         Environments
// @Produce      json
// @Param        org_id     path int    true "Organization ID"
// @Param        project_id path int    true "Project ID"
// @Param        env_name   path string true "Environment name"
// @Success      200 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Router       /organizations/{org_id}/projects/{project_id}/environments/{env_name} [delete]
func DeleteEnvironment(c fiber.Ctx) (int, response.APIResponse) {
	proj, status, errResp := resolveProject(c)
	if errResp != nil {
		return status, errResp
	}
	if status, errResp := requirePermission(proj.OrganizationID, permissions.EnvDelete, c); errResp != nil {
		return status, errResp
	}
	envs, err := controllers.DeleteEnvironment(proj, c.Params("env_name"))
	if err != nil {
		return envErr(err)
	}
	return fiber.StatusOK, response.DataResponse{Data: envs}
}
