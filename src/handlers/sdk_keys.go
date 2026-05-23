package handlers

import (
	"errors"
	"strconv"

	"flagpole/src/controllers"
	"flagpole/src/models"
	"flagpole/src/pkg/permissions"
	"flagpole/src/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type sdkKeyCreateRequest struct {
	Name          string `json:"name"`
	KeyType       string `json:"type"`
	EnvironmentID uint   `json:"environmentId"`
}

func sdkKeyErr(err error) (int, response.APIResponse) {
	switch {
	case errors.Is(err, controllers.ErrSDKKeyNotFound),
		errors.Is(err, controllers.ErrSDKKeyEnvNotFound):
		return fiber.StatusNotFound, response.ErrorResponse{Error: err.Error()}
	case errors.Is(err, controllers.ErrSDKKeyNameRequired),
		errors.Is(err, controllers.ErrSDKKeyNameTooLong),
		errors.Is(err, controllers.ErrSDKKeyTypeInvalid):
		return fiber.StatusBadRequest, response.ErrorResponse{Error: err.Error()}
	case errors.Is(err, controllers.ErrSDKKeyLimitReached):
		return fiber.StatusUnprocessableEntity, response.ErrorResponse{Error: err.Error()}
	case errors.Is(err, controllers.ErrSDKKeyAlreadyRevoked):
		return fiber.StatusConflict, response.ErrorResponse{Error: err.Error()}
	default:
		return fiber.StatusInternalServerError, response.Error500
	}
}

// ListSDKKeys godoc
// @Summary      List SDK keys for a project
// @Tags         SDK Keys
// @Produce      json
// @Param        org_id      path  int  true  "Organization ID"
// @Param        project_id  path  int  true  "Project ID"
// @Param        env_id      query int  false "Filter by environment ID"
// @Success      200 {object} response.DataResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Router       /organizations/{org_id}/projects/{project_id}/sdk-keys [get]
func ListSDKKeys(c fiber.Ctx) (int, response.APIResponse) {
	proj, status, errResp := resolveProject(c)
	if errResp != nil {
		return status, errResp
	}
	if status, errResp := requirePermission(proj.OrganizationID, permissions.SDKView, c); errResp != nil {
		return status, errResp
	}
	var envID uint
	if raw := c.Query("env_id"); raw != "" {
		v, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid env_id"}
		}
		envID = uint(v)
	}
	keys, err := controllers.ListSDKKeys(proj.ID, envID)
	if err != nil {
		return fiber.StatusInternalServerError, response.Error500
	}
	return fiber.StatusOK, response.DataResponse{Data: keys}
}

// CreateSDKKey godoc
// @Summary      Create an SDK key
// @Tags         SDK Keys
// @Accept       json
// @Produce      json
// @Param        org_id      path int  true "Organization ID"
// @Param        project_id  path int  true "Project ID"
// @Param        body        body sdkKeyCreateRequest true "Key definition"
// @Success      201 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      422 {object} response.ErrorResponse
// @Router       /organizations/{org_id}/projects/{project_id}/sdk-keys [post]
func CreateSDKKey(c fiber.Ctx) (int, response.APIResponse) {
	proj, status, errResp := resolveProject(c)
	if errResp != nil {
		return status, errResp
	}
	if status, errResp := requirePermission(proj.OrganizationID, permissions.SDKCreate, c); errResp != nil {
		return status, errResp
	}
	var req sdkKeyCreateRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "couldn't parse body"}
	}
	created, err := controllers.CreateSDKKey(proj.ID, req.EnvironmentID, req.KeyType, req.Name)
	if err != nil {
		return sdkKeyErr(err)
	}
	logAudit(c, proj.OrganizationID, &proj.ID, models.ActionSDKKeyCreate, created.EnvironmentName, "Created "+req.KeyType+" SDK key '"+req.Name+"'", created.EnvironmentName)
	return fiber.StatusCreated, response.DataResponse{Data: created}
}

// RevokeSDKKey godoc
// @Summary      Revoke an SDK key
// @Tags         SDK Keys
// @Produce      json
// @Param        org_id      path int true "Organization ID"
// @Param        project_id  path int true "Project ID"
// @Param        key_id      path int true "SDK Key ID"
// @Success      204
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      409 {object} response.ErrorResponse
// @Router       /organizations/{org_id}/projects/{project_id}/sdk-keys/{key_id} [delete]
func RevokeSDKKey(c fiber.Ctx) (int, response.APIResponse) {
	proj, status, errResp := resolveProject(c)
	if errResp != nil {
		return status, errResp
	}
	if status, errResp := requirePermission(proj.OrganizationID, permissions.SDKRevoke, c); errResp != nil {
		return status, errResp
	}
	keyID, err := strconv.ParseUint(c.Params("key_id"), 10, 64)
	if err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid key id"}
	}
	keyName, err := controllers.RevokeSDKKey(uint(keyID), proj.ID)
	if err != nil {
		return sdkKeyErr(err)
	}
	logAudit(c, proj.OrganizationID, &proj.ID, models.ActionSDKKeyRevoke, keyName, "Revoked SDK key '"+keyName+"'", "")
	return fiber.StatusNoContent, nil
}

// RevealSDKKey godoc
// @Summary      Reveal the raw value of an SDK key
// @Tags         SDK Keys
// @Produce      json
// @Param        org_id      path int true "Organization ID"
// @Param        project_id  path int true "Project ID"
// @Param        key_id      path int true "SDK Key ID"
// @Success      200 {object} response.DataResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      409 {object} response.ErrorResponse
// @Router       /organizations/{org_id}/projects/{project_id}/sdk-keys/{key_id}/reveal [get]
func RevealSDKKey(c fiber.Ctx) (int, response.APIResponse) {
	proj, status, errResp := resolveProject(c)
	if errResp != nil {
		return status, errResp
	}
	if status, errResp := requirePermission(proj.OrganizationID, permissions.SDKView, c); errResp != nil {
		return status, errResp
	}
	keyID, err := strconv.ParseUint(c.Params("key_id"), 10, 64)
	if err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid key id"}
	}
	rawKey, keyName, err := controllers.RevealSDKKey(uint(keyID), proj.ID)
	if err != nil {
		return sdkKeyErr(err)
	}
	logAudit(c, proj.OrganizationID, &proj.ID, models.ActionSDKKeyReveal, keyName, "Revealed SDK key '"+keyName+"'", "")
	return fiber.StatusOK, response.DataResponse{Data: map[string]string{"key": rawKey}}
}
