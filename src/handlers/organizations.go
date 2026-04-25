package handlers

import (
	"strconv"

	"flagpole/src/controllers"
	"flagpole/src/dal"
	"flagpole/src/pkg/jwtutil"
	"flagpole/src/pkg/response"

	"github.com/gofiber/fiber/v3"
)

const internalOrgName = "Flagpole"

type orgRequest struct {
	Name string `json:"name"`
}

func isInternalUser(c fiber.Ctx) bool {
	orgNames, ok := jwtutil.Claims(c)["orgNames"].([]interface{})
	if !ok {
		return false
	}
	for _, name := range orgNames {
		if name == internalOrgName {
			return true
		}
	}
	return false
}

// ListOrganizations godoc
// @Summary      List all organizations
// @Tags         Organizations
// @Produce      json
// @Success      200 {object} response.DataResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /organizations [get]
func ListOrganizations(c fiber.Ctx) (int, response.APIResponse) {
	if !isInternalUser(c) {
		return fiber.StatusForbidden, response.ErrorResponse{Error: "forbidden"}
	}

	orgs, err := dal.Organization.List()
	if err != nil {
		return fiber.StatusInternalServerError, response.Error500
	}

	return fiber.StatusOK, response.DataResponse{Data: orgs}
}

// GetOrganization godoc
// @Summary      Get an organization by ID
// @Tags         Organizations
// @Produce      json
// @Param        id path int true "Organization ID"
// @Success      200 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Router       /organizations/{id} [get]
func GetOrganization(c fiber.Ctx) (int, response.APIResponse) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid id"}
	}

	org, err := dal.Organization.GetByID(uint(id))
	if err != nil {
		return fiber.StatusNotFound, response.ErrorResponse{Error: "organization not found"}
	}

	if org.Name == internalOrgName && !isInternalUser(c) {
		return fiber.StatusNotFound, response.ErrorResponse{Error: "organization not found"}
	}

	return fiber.StatusOK, response.DataResponse{Data: org}
}

// CreateOrganization godoc
// @Summary      Create an organization
// @Tags         Organizations
// @Accept       json
// @Produce      json
// @Param        body body orgRequest true "Organization data"
// @Success      201 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Router       /organizations [post]
func CreateOrganization(c fiber.Ctx) (int, response.APIResponse) {
	var req orgRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "couldn't parse body"}
	}
	if req.Name == "" {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "name is required"}
	}
	if req.Name == internalOrgName {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid organization name"}
	}

	org, err := controllers.CreateOrganization(req.Name, jwtutil.UserID(c))
	if err == controllers.ErrOrgLimitReached {
		return fiber.StatusForbidden, response.ErrOrgLimitReached
	}
	if err != nil {
		return fiber.StatusInternalServerError, response.Error500
	}

	return fiber.StatusCreated, response.DataResponse{Data: org}
}

// UpdateOrganization godoc
// @Summary      Update an organization
// @Tags         Organizations
// @Accept       json
// @Produce      json
// @Param        id   path int        true "Organization ID"
// @Param        body body orgRequest true "Organization data"
// @Success      200 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Router       /organizations/{id} [put]
func UpdateOrganization(c fiber.Ctx) (int, response.APIResponse) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid id"}
	}

	org, err := dal.Organization.GetByID(uint(id))
	if err != nil {
		return fiber.StatusNotFound, response.ErrorResponse{Error: "organization not found"}
	}

	if org.Name == internalOrgName && !isInternalUser(c) {
		return fiber.StatusNotFound, response.ErrorResponse{Error: "organization not found"}
	}

	var req orgRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "couldn't parse body"}
	}
	if req.Name == "" {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "name is required"}
	}

	org.Name = req.Name
	if err := dal.Organization.Save(org); err != nil {
		return fiber.StatusInternalServerError, response.Error500
	}

	return fiber.StatusOK, response.DataResponse{Data: org}
}

// DeleteOrganization godoc
// @Summary      Delete an organization
// @Tags         Organizations
// @Param        id path int true "Organization ID"
// @Success      204
// @Failure      400 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Router       /organizations/{id} [delete]
func DeleteOrganization(c fiber.Ctx) (int, response.APIResponse) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid id"}
	}

	org, err := dal.Organization.GetByID(uint(id))
	if err != nil {
		return fiber.StatusNotFound, response.ErrorResponse{Error: "organization not found"}
	}

	if org.Name == internalOrgName && !isInternalUser(c) {
		return fiber.StatusNotFound, response.ErrorResponse{Error: "organization not found"}
	}

	if err := dal.Organization.Delete(org); err != nil {
		return fiber.StatusInternalServerError, response.Error500
	}

	return fiber.StatusNoContent, nil
}
