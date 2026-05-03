package handlers

import (
	"errors"
	"fmt"
	"strconv"

	"flagpole/src/controllers"
	"flagpole/src/models"
	"flagpole/src/pkg/permissions"
	"flagpole/src/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type flagCreateRequest struct {
	Key         string      `json:"key"`
	Description string      `json:"description"`
	FlagType    string      `json:"type"`
	Value       interface{} `json:"value"`
}

type flagUpdateRequest struct {
	Description *string     `json:"description"`
	Value       interface{} `json:"value"`
	Enabled     *bool       `json:"enabled"`
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
	if status, errResp := requirePermission(proj.OrganizationID, permissions.FlagCreate, c); errResp != nil {
		return status, errResp
	}
	flag, err := controllers.CreateFlag(proj.ID, req.Key, req.Description, models.FlagType(req.FlagType), req.Value)
	if err != nil {
		return flagErr(err)
	}
	logAudit(c, proj.OrganizationID, &proj.ID, models.ActionFlagCreate, flag.Key, "Created flag '"+flag.Key+"'", "")
	return fiber.StatusCreated, response.DataResponse{Data: flag}
}

// GetFlag godoc
// @Summary      Get a feature flag by ID with environment config
// @Tags         Flags
// @Produce      json
// @Param        org_id     path int true "Organization ID"
// @Param        project_id path int true "Project ID"
// @Param        flag_id    path int true "Flag ID"
// @Param        env        query string true "Environment name"
// @Success      200 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Router       /organizations/{org_id}/projects/{project_id}/flags/{flag_id} [get]
func GetFlag(c fiber.Ctx) (int, response.APIResponse) {
	proj, status, errResp := resolveProject(c)
	if errResp != nil {
		return status, errResp
	}
	flagID, err := strconv.ParseUint(c.Params("flag_id"), 10, 64)
	if err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid flag id"}
	}
	env := c.Query("env")
	if env == "" {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "env query parameter is required"}
	}
	detail, err := controllers.GetFlagDetail(proj.ID, uint(flagID), env)
	if err != nil {
		return flagErr(err)
	}
	return fiber.StatusOK, response.DataResponse{Data: detail}
}

// UpdateFlag godoc
// @Summary      Update a feature flag (description only)
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
	flag, proj, status, errResp := resolveFlag(c)
	if errResp != nil {
		return status, errResp
	}
	var req flagUpdateRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "couldn't parse body"}
	}
	if status, errResp := requirePermission(proj.OrganizationID, permissions.FlagUpdate, c); errResp != nil {
		return status, errResp
	}
	if req.Description != nil {
		updated, err := controllers.UpdateFlagMetadata(flag, req.Description)
		if err != nil {
			return flagErr(err)
		}
		flag = updated
	}
	logAudit(c, proj.OrganizationID, &proj.ID, models.ActionFlagUpdate, flag.Key, "Updated flag '"+flag.Key+"'", "")
	return fiber.StatusOK, response.DataResponse{Data: flag}
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
	flag, proj, status, errResp := resolveFlag(c)
	if errResp != nil {
		return status, errResp
	}
	if status, errResp := requirePermission(proj.OrganizationID, permissions.FlagDelete, c); errResp != nil {
		return status, errResp
	}
	if err := controllers.DeleteFlag(flag); err != nil {
		return fiber.StatusInternalServerError, response.Error500
	}
	logAudit(c, proj.OrganizationID, &proj.ID, models.ActionFlagDelete, flag.Key, "Deleted flag '"+flag.Key+"'", "")
	return fiber.StatusNoContent, nil
}

// GetFlagAudit godoc
// @Summary      Get audit log for a specific flag
// @Tags         Flags
// @Produce      json
// @Param        org_id     path int true "Organization ID"
// @Param        project_id path int true "Project ID"
// @Param        flag_id    path int true "Flag ID"
// @Param        env        query string true "Environment name"
// @Success      200 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Router       /organizations/{org_id}/projects/{project_id}/flags/{flag_id}/audit [get]
func GetFlagAudit(c fiber.Ctx) (int, response.APIResponse) {
	flag, proj, status, errResp := resolveFlag(c)
	if errResp != nil {
		return status, errResp
	}
	env := c.Query("env")
	audits, err := controllers.GetFlagAudit(proj.ID, flag.Key, env)
	if err != nil {
		return fiber.StatusInternalServerError, response.Error500
	}
	return fiber.StatusOK, response.DataResponse{Data: audits}
}

type flagConfigRequest struct {
	Enabled           *bool                    `json:"enabled"`
	RolloutEnabled    *bool                    `json:"rolloutEnabled"`
	RolloutPercentage *int                     `json:"rolloutPercentage"`
	DefaultValue      interface{}              `json:"defaultValue"`
	ServedValue       interface{}              `json:"servedValue"`
	Overrides         []controllers.OverridePayload `json:"overrides"`
}

// CreateFlagEnvConfig godoc
// @Summary      Create flag configuration for an environment
// @Tags         Flags
// @Accept       json
// @Produce      json
// @Param        org_id     path int true "Organization ID"
// @Param        project_id path int true "Project ID"
// @Param        flag_id    path int true "Flag ID"
// @Param        env        query string true "Environment name"
// @Success      201 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      409 {object} response.ErrorResponse
// @Router       /organizations/{org_id}/projects/{project_id}/flags/{flag_id}/config [post]
func CreateFlagEnvConfig(c fiber.Ctx) (int, response.APIResponse) {
	flag, proj, status, errResp := resolveFlag(c)
	if errResp != nil {
		return status, errResp
	}
	env := c.Query("env")
	if env == "" {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "env query parameter is required"}
	}
	if status, errResp := requirePermission(proj.OrganizationID, permissions.FlagUpdate, c); errResp != nil {
		return status, errResp
	}
	config, err := controllers.CreateFlagEnvConfig(flag.ID, env, proj.ID, flag.FlagType)
	if err != nil {
		if errors.Is(err, controllers.ErrConfigExists) {
			return fiber.StatusConflict, response.ErrorResponse{Error: "configuration already exists for this environment"}
		}
		return flagErr(err)
	}
	logAudit(c, proj.OrganizationID, &proj.ID, models.ActionFlagUpdate, flag.Key, "Created configuration for '"+env+"'", env)
	return fiber.StatusCreated, response.DataResponse{Data: config}
}

// UpdateFlagConfig godoc
// @Summary      Update flag configuration for an environment
// @Tags         Flags
// @Accept       json
// @Produce      json
// @Param        org_id     path int true "Organization ID"
// @Param        project_id path int true "Project ID"
// @Param        flag_id    path int true "Flag ID"
// @Param        env        query string true "Environment name"
// @Param        body       body flagConfigRequest true "Configuration"
// @Success      200 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Router       /organizations/{org_id}/projects/{project_id}/flags/{flag_id}/config [patch]
func UpdateFlagConfig(c fiber.Ctx) (int, response.APIResponse) {
	flag, proj, status, errResp := resolveFlag(c)
	if errResp != nil {
		return status, errResp
	}
	env := c.Query("env")
	if env == "" {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "env query parameter is required"}
	}
	var req flagConfigRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "couldn't parse body"}
	}
	if status, errResp := requirePermission(proj.OrganizationID, permissions.FlagUpdate, c); errResp != nil {
		return status, errResp
	}

	flagType := models.FlagType(flag.FlagType)
	if req.DefaultValue != nil {
		if err := models.ValidateValue(flagType, req.DefaultValue); err != nil {
			return fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid defaultValue: " + err.Error()}
		}
	}
	if req.ServedValue != nil {
		if err := models.ValidateValue(flagType, req.ServedValue); err != nil {
			return fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid servedValue: " + err.Error()}
		}
	}
	for _, o := range req.Overrides {
		if err := models.ValidateValue(flagType, o.Value); err != nil {
			return fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid override value: " + err.Error()}
		}
	}

	changes, err := controllers.UpdateFlagConfig(flag.ID, env, req.Enabled, req.RolloutEnabled, req.RolloutPercentage, req.DefaultValue, req.ServedValue, req.Overrides)
	if err != nil {
		if errors.Is(err, controllers.ErrConfigNotFound) {
			return fiber.StatusNotFound, response.ErrorResponse{Error: "configuration not found for this environment"}
		}
		return flagErr(err)
	}

	if changes.EnabledChanged != nil {
		state := "Disabled"
		if *changes.EnabledChanged {
			state = "Enabled"
		}
		logAudit(c, proj.OrganizationID, &proj.ID, models.ActionFlagToggle, flag.Key, state+" flag '"+flag.Key+"' in "+env, env)
	}
	if changes.RolloutToggled != nil {
		state := "Disabled"
		if *changes.RolloutToggled {
			state = "Enabled"
		}
		logAudit(c, proj.OrganizationID, &proj.ID, models.ActionFlagRollout, flag.Key, state+" rollout for '"+flag.Key+"' in "+env, env)
	}
	if changes.RolloutPctChanged {
		logAudit(c, proj.OrganizationID, &proj.ID, models.ActionFlagRollout, flag.Key, "Rollout for '"+flag.Key+"' set to "+fmt.Sprint(changes.RolloutPct)+"% in "+env, env)
	}
	if changes.ValuesChanged {
		logAudit(c, proj.OrganizationID, &proj.ID, models.ActionFlagValues, flag.Key, "Updated values for '"+flag.Key+"' in "+env, env)
	}
	for _, segID := range changes.OverridesAdded {
		_ = segID
		logAudit(c, proj.OrganizationID, &proj.ID, models.ActionFlagOverrideAdd, flag.Key, "Added segment override for '"+flag.Key+"' in "+env, env)
	}
	for _, segID := range changes.OverridesRemoved {
		_ = segID
		logAudit(c, proj.OrganizationID, &proj.ID, models.ActionFlagOverrideRemove, flag.Key, "Removed segment override for '"+flag.Key+"' in "+env, env)
	}

	return fiber.StatusOK, response.DataResponse{Data: fiber.Map{"ok": true}}
}
