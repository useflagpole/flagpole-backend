package handlers

import (
	"errors"
	"strconv"

	"flagpole/src/controllers"
	"flagpole/src/models"
	"flagpole/src/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type flagCreateRequest struct {
	Key      string      `json:"key"`
	Name     string      `json:"name"`
	FlagType string      `json:"type"`
	Value    interface{} `json:"value"`
}

type flagUpdateRequest struct {
	Name    *string     `json:"name"`
	Value   interface{} `json:"value"`
	Enabled *bool       `json:"enabled"`
}

func flagErr(err error) (int, response.APIResponse) {
	switch {
	case errors.Is(err, controllers.ErrFlagKeyInvalid):
		return fiber.StatusBadRequest, response.ErrorResponse{Error: err.Error()}
	case errors.Is(err, controllers.ErrFlagKeyTaken):
		return fiber.StatusConflict, response.ErrorResponse{Error: err.Error()}
	case errors.Is(err, controllers.ErrFlagLimitReached):
		return fiber.StatusUnprocessableEntity, response.ErrorResponse{Error: err.Error()}
	case errors.Is(err, controllers.ErrFlagNotFound):
		return fiber.StatusNotFound, response.ErrorResponse{Error: err.Error()}
	case errors.Is(err, controllers.ErrFlagNameRequired),
		errors.Is(err, controllers.ErrFlagNameInvalid):
		return fiber.StatusBadRequest, response.ErrorResponse{Error: err.Error()}
	default:
		return fiber.StatusBadRequest, response.ErrorResponse{Error: err.Error()}
	}
}

func resolveFlag(c fiber.Ctx) (*models.FeatureFlag, *models.Project, int, response.APIResponse) {
	proj, status, errResp := resolveProject(c)
	if errResp != nil {
		return nil, nil, status, errResp
	}
	flagID, err := strconv.ParseUint(c.Params("flag_id"), 10, 64)
	if err != nil {
		return nil, nil, fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid flag id"}
	}
	flag, err := controllers.GetFlag(proj.ID, uint(flagID))
	if err != nil {
		return nil, nil, fiber.StatusNotFound, response.ErrorResponse{Error: err.Error()}
	}
	return flag, proj, 0, nil
}

// ListFlags godoc
// @Summary      List feature flags for a project
// @Tags         Flags
// @Produce      json
// @Param        org_id     path int true "Organization ID"
// @Param        project_id path int true "Project ID"
// @Success      200 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Router       /organizations/{org_id}/projects/{project_id}/flags [get]
func ListFlags(c fiber.Ctx) (int, response.APIResponse) {
	proj, status, errResp := resolveProject(c)
	if errResp != nil {
		return status, errResp
	}
	flags, err := controllers.ListFlags(proj.ID)
	if err != nil {
		return fiber.StatusInternalServerError, response.Error500
	}
	return fiber.StatusOK, response.DataResponse{Data: flags}
}

// CreateFlag godoc
// @Summary      Create a feature flag
// @Tags         Flags
// @Accept       json
// @Produce      json
// @Param        org_id     path int true "Organization ID"
// @Param        project_id path int true "Project ID"
// @Param        body       body flagCreateRequest true "Flag definition"
// @Success      201 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      409 {object} response.ErrorResponse
// @Router       /organizations/{org_id}/projects/{project_id}/flags [post]
func CreateFlag(c fiber.Ctx) (int, response.APIResponse) {
	proj, status, errResp := resolveProject(c)
	if errResp != nil {
		return status, errResp
	}
	var req flagCreateRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "couldn't parse body"}
	}
	flag, err := controllers.CreateFlag(proj.ID, req.Key, req.Name, models.FlagType(req.FlagType), req.Value)
	if err != nil {
		return flagErr(err)
	}
	return fiber.StatusCreated, response.DataResponse{Data: flag}
}

// GetFlag godoc
// @Summary      Get a feature flag by ID
// @Tags         Flags
// @Produce      json
// @Param        org_id     path int true "Organization ID"
// @Param        project_id path int true "Project ID"
// @Param        flag_id    path int true "Flag ID"
// @Success      200 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Router       /organizations/{org_id}/projects/{project_id}/flags/{flag_id} [get]
func GetFlag(c fiber.Ctx) (int, response.APIResponse) {
	flag, _, status, errResp := resolveFlag(c)
	if errResp != nil {
		return status, errResp
	}
	return fiber.StatusOK, response.DataResponse{Data: flag}
}

// UpdateFlag godoc
// @Summary      Update a feature flag (name, value, enabled)
// @Tags         Flags
// @Accept       json
// @Produce      json
// @Param        org_id     path int true "Organization ID"
// @Param        project_id path int true "Project ID"
// @Param        flag_id    path int true "Flag ID"
// @Param        body       body flagUpdateRequest true "Fields to update"
// @Success      200 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Router       /organizations/{org_id}/projects/{project_id}/flags/{flag_id} [patch]
func UpdateFlag(c fiber.Ctx) (int, response.APIResponse) {
	flag, _, status, errResp := resolveFlag(c)
	if errResp != nil {
		return status, errResp
	}
	var req flagUpdateRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "couldn't parse body"}
	}
	updated, err := controllers.UpdateFlag(flag, req.Name, req.Value, req.Enabled)
	if err != nil {
		return flagErr(err)
	}
	return fiber.StatusOK, response.DataResponse{Data: updated}
}

// DeleteFlag godoc
// @Summary      Delete a feature flag
// @Tags         Flags
// @Produce      json
// @Param        org_id     path int true "Organization ID"
// @Param        project_id path int true "Project ID"
// @Param        flag_id    path int true "Flag ID"
// @Success      204
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /organizations/{org_id}/projects/{project_id}/flags/{flag_id} [delete]
func DeleteFlag(c fiber.Ctx) (int, response.APIResponse) {
	flag, _, status, errResp := resolveFlag(c)
	if errResp != nil {
		return status, errResp
	}
	if err := controllers.DeleteFlag(flag); err != nil {
		return fiber.StatusInternalServerError, response.Error500
	}
	return fiber.StatusNoContent, nil
}
