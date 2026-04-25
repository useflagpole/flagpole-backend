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

type envNameRequest struct {
	Name string `json:"name"`
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

func CreateEnvironment(c fiber.Ctx) (int, response.APIResponse) {
	proj, status, errResp := resolveProject(c)
	if errResp != nil {
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

func RenameEnvironment(c fiber.Ctx) (int, response.APIResponse) {
	proj, status, errResp := resolveProject(c)
	if errResp != nil {
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

func DeleteEnvironment(c fiber.Ctx) (int, response.APIResponse) {
	proj, status, errResp := resolveProject(c)
	if errResp != nil {
		return status, errResp
	}
	envs, err := controllers.DeleteEnvironment(proj, c.Params("env_name"))
	if err != nil {
		return envErr(err)
	}
	return fiber.StatusOK, response.DataResponse{Data: envs}
}
