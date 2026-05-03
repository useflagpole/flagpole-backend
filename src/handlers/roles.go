package handlers

import (
	"strconv"
	"strings"

	"flagpole/src/dal"
	"flagpole/src/models"
	"flagpole/src/pkg/jwtutil"
	"flagpole/src/pkg/permissions"
	"flagpole/src/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type roleWithPermissions struct {
	ID          uint     `json:"id"`
	Name        string   `json:"name"`
	IsDefault   bool     `json:"isDefault"`
	IsProtected bool     `json:"isProtected"`
	Permissions []string `json:"permissions"`
}

type createRoleRequest struct {
	Name string `json:"name"`
}

type updatePermissionsRequest struct {
	Permissions []string `json:"permissions"`
}

func parseOrgID(c fiber.Ctx) (uint, error) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	return uint(id), err
}

// ListOrgRoles godoc
// @Summary      List roles for an organization
// @Tags         Roles
// @Produce      json
// @Param        id path int true "Organization ID"
// @Success      200 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Router       /organizations/{id}/roles [get]
func ListOrgRoles(c fiber.Ctx) (int, response.APIResponse) {
	orgID, err := parseOrgID(c)
	if err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid id"}
	}
	if !dal.Organization.IsMember(orgID, jwtutil.UserID(c)) && !isInternalUser(c) {
		return fiber.StatusForbidden, response.ErrorResponse{Error: "forbidden"}
	}

	roles, err := dal.OrgRole.ListByOrg(orgID)
	if err != nil {
		return fiber.StatusInternalServerError, response.Error500
	}

	out := make([]roleWithPermissions, 0, len(roles))
	for _, r := range roles {
		perms, err := dal.OrgRole.GetPermissions(r.ID)
		if err != nil {
			return fiber.StatusInternalServerError, response.Error500
		}
		out = append(out, roleWithPermissions{
			ID:          r.ID,
			Name:        r.Name,
			IsDefault:   r.IsDefault,
			IsProtected: r.IsProtected,
			Permissions: perms,
		})
	}

	return fiber.StatusOK, response.DataResponse{Data: out}
}

// CreateOrgRole godoc
// @Summary      Create a role for an organization
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Param        id   path int              true "Organization ID"
// @Param        body body createRoleRequest true "Role data"
// @Success      201 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Router       /organizations/{id}/roles [post]
func CreateOrgRole(c fiber.Ctx) (int, response.APIResponse) {
	orgID, err := parseOrgID(c)
	if err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid id"}
	}
	if status, errResp := requirePermission(orgID, permissions.OrgRoles, c); errResp != nil {
		return status, errResp
	}

	var req createRoleRequest
	if err := c.Bind().JSON(&req); err != nil || req.Name == "" {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "name is required"}
	}

	name := strings.ToLower(strings.TrimSpace(req.Name))
	if len(name) < 2 {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "role name must be at least 2 characters"}
	}

	role := &models.OrgRole{
		OrganizationID: orgID,
		Name:           name,
		IsDefault:      false,
		IsProtected:    false,
	}
	if err := dal.OrgRole.Create(role); err != nil {
		return fiber.StatusInternalServerError, response.Error500
	}

	return fiber.StatusCreated, response.DataResponse{Data: roleWithPermissions{
		ID:          role.ID,
		Name:        role.Name,
		IsDefault:   role.IsDefault,
		IsProtected: role.IsProtected,
		Permissions: []string{},
	}}
}

// DeleteOrgRole godoc
// @Summary      Delete a role from an organization
// @Tags         Roles
// @Param        id     path int true "Organization ID"
// @Param        roleId path int true "Role ID"
// @Success      204
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Router       /organizations/{id}/roles/{roleId} [delete]
func DeleteOrgRole(c fiber.Ctx) (int, response.APIResponse) {
	orgID, err := parseOrgID(c)
	if err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid id"}
	}
	if status, errResp := requirePermission(orgID, permissions.OrgRoles, c); errResp != nil {
		return status, errResp
	}

	roleID, err := strconv.ParseUint(c.Params("roleId"), 10, 64)
	if err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid role id"}
	}

	role, err := dal.OrgRole.GetByID(uint(roleID))
	if err != nil {
		return fiber.StatusNotFound, response.ErrorResponse{Error: "role not found"}
	}
	if role.OrganizationID != orgID {
		return fiber.StatusNotFound, response.ErrorResponse{Error: "role not found"}
	}

	if err := dal.OrgRole.Delete(role); err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: err.Error()}
	}

	return fiber.StatusNoContent, nil
}

// UpdateOrgRolePermissions godoc
// @Summary      Replace all permissions for a role
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Param        id     path int                      true "Organization ID"
// @Param        roleId path int                      true "Role ID"
// @Param        body   body updatePermissionsRequest true "Permissions list"
// @Success      200 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Router       /organizations/{id}/roles/{roleId}/permissions [put]
func UpdateOrgRolePermissions(c fiber.Ctx) (int, response.APIResponse) {
	orgID, err := parseOrgID(c)
	if err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid id"}
	}
	if status, errResp := requirePermission(orgID, permissions.OrgRoles, c); errResp != nil {
		return status, errResp
	}

	roleID, err := strconv.ParseUint(c.Params("roleId"), 10, 64)
	if err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid role id"}
	}

	role, err := dal.OrgRole.GetByID(uint(roleID))
	if err != nil {
		return fiber.StatusNotFound, response.ErrorResponse{Error: "role not found"}
	}
	if role.OrganizationID != orgID {
		return fiber.StatusNotFound, response.ErrorResponse{Error: "role not found"}
	}
	if role.IsProtected {
		return fiber.StatusForbidden, response.ErrorResponse{Error: "cannot modify protected role"}
	}

	var req updatePermissionsRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "couldn't parse body"}
	}

	// Build desired set
	desired := make(map[string]bool, len(req.Permissions))
	for _, p := range req.Permissions {
		desired[p] = true
	}

	// Build valid set from all known permissions
	valid := make(map[string]bool, len(permissions.All))
	for _, p := range permissions.All {
		valid[p.Code] = true
	}

	current, err := dal.OrgRole.GetPermissions(uint(roleID))
	if err != nil {
		return fiber.StatusInternalServerError, response.Error500
	}
	currentSet := make(map[string]bool, len(current))
	for _, p := range current {
		currentSet[p] = true
	}

	// Enable new perms
	for _, p := range req.Permissions {
		if !valid[p] {
			continue
		}
		if !currentSet[p] {
			if err := dal.OrgRole.SetPermission(uint(roleID), p, true); err != nil {
				return fiber.StatusInternalServerError, response.Error500
			}
		}
	}
	// Disable removed perms
	for _, p := range current {
		if !desired[p] {
			if err := dal.OrgRole.SetPermission(uint(roleID), p, false); err != nil {
				return fiber.StatusInternalServerError, response.Error500
			}
		}
	}

	updated, err := dal.OrgRole.GetPermissions(uint(roleID))
	if err != nil {
		return fiber.StatusInternalServerError, response.Error500
	}

	return fiber.StatusOK, response.DataResponse{Data: roleWithPermissions{
		ID:          role.ID,
		Name:        role.Name,
		IsDefault:   role.IsDefault,
		IsProtected: role.IsProtected,
		Permissions: updated,
	}}
}
