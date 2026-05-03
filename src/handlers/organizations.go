package handlers

import (
	"strconv"

	"flagpole/src/controllers"
	"flagpole/src/dal"
	"flagpole/src/models"
	"flagpole/src/pkg/jwtutil"
	"flagpole/src/pkg/permissions"
	"flagpole/src/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type orgRequest struct {
	Name string `json:"name"`
}

type planRequest struct {
	Plan string `json:"plan"`
}

type updateMemberRoleRequest struct {
	RoleID uint `json:"roleId"`
}

func isInternalUser(c fiber.Ctx) bool {
	orgs, err := dal.Organization.ListByUser(jwtutil.UserID(c))
	if err != nil {
		return false
	}
	for _, o := range orgs {
		if controllers.IsInternalOrg(o.Name) {
			return true
		}
	}
	return false
}

// ListOrganizations godoc
// @Summary      List all organizations, can only be accessed by internal users
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

	isMember := dal.Organization.IsMember(uint(id), jwtutil.UserID(c))
	if !isMember && !isInternalUser(c) {
		return fiber.StatusForbidden, response.ErrorResponse{Error: "forbidden"}
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

	org, err := controllers.CreateOrganization(req.Name, jwtutil.UserID(c))
	switch err {
	case nil:
	case controllers.ErrReservedOrgName:
		return fiber.StatusBadRequest, response.ErrorResponse{Error: err.Error()}
	case controllers.ErrOrgLimitReached:
		return fiber.StatusForbidden, response.ErrOrgLimitReached
	default:
		return fiber.StatusInternalServerError, response.Error500
	}

	logAudit(c, org.ID, nil, models.ActionOrgCreate, org.Name, "Created organization '"+org.Name+"'", "")
	return fiber.StatusCreated, response.DataResponse{Data: org}
}

// SetOrganizationPlan godoc
// @Summary      Set an organization's plan
// @Tags         Organizations
// @Accept       json
// @Produce      json
// @Param        id   path int        true "Organization ID"
// @Param        body body planRequest true "Plan"
// @Success      200 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Router       /organizations/{id}/plan [patch]
func SetOrganizationPlan(c fiber.Ctx) (int, response.APIResponse) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid id"}
	}

	if _, err := dal.Organization.GetByID(uint(id)); err != nil {
		return fiber.StatusNotFound, response.ErrorResponse{Error: "organization not found"}
	}

	var req planRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "couldn't parse body"}
	}
	if req.Plan == "" {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "plan is required"}
	}

	if err := controllers.SetOrganizationPlan(uint(id), req.Plan); err == controllers.ErrInvalidPlan {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid plan"}
	} else if err != nil {
		return fiber.StatusInternalServerError, response.Error500
	}

	logAudit(c, uint(id), nil, models.ActionOrgPlan, req.Plan, "Set organization plan to '"+req.Plan+"'", "")

	return fiber.StatusOK, response.DataResponse{Data: fiber.Map{"plan": req.Plan}}
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

	if controllers.IsInternalOrg(org.Name) && !isInternalUser(c) {
		return fiber.StatusNotFound, response.ErrorResponse{Error: "organization not found"}
	}
	if status, errResp := requirePermission(uint(id), permissions.OrgRename, c); errResp != nil {
		return status, errResp
	}

	var req orgRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "couldn't parse body"}
	}
	if req.Name == "" {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "name is required"}
	}

	oldName := org.Name
	updated, err := controllers.UpdateOrganization(org, req.Name)
	if err != nil {
		return fiber.StatusInternalServerError, response.Error500
	}

	logAudit(c, updated.ID, nil, models.ActionOrgRename, updated.Name, "Renamed organization from '"+oldName+"' to '"+updated.Name+"'", "")
	return fiber.StatusOK, response.DataResponse{Data: updated}
}

// ListOrgMembers godoc
// @Summary      List members of an organization with their roles
// @Tags         Organizations
// @Produce      json
// @Param        id path int true "Organization ID"
// @Success      200 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Router       /organizations/{id}/members [get]
func ListOrgMembers(c fiber.Ctx) (int, response.APIResponse) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid id"}
	}
	if _, err := dal.Organization.GetByID(uint(id)); err != nil {
		return fiber.StatusNotFound, response.ErrorResponse{Error: "organization not found"}
	}
	if !dal.Organization.IsMember(uint(id), jwtutil.UserID(c)) && !isInternalUser(c) {
		return fiber.StatusForbidden, response.ErrorResponse{Error: "forbidden"}
	}
	members, err := dal.Organization.ListMembers(uint(id))
	if err != nil {
		return fiber.StatusInternalServerError, response.Error500
	}
	return fiber.StatusOK, response.DataResponse{Data: members}
}

// UpdateMemberRole godoc
// @Summary      Update a member's role in an organization
// @Tags         Organizations
// @Accept       json
// @Produce      json
// @Param        id     path int                     true "Organization ID"
// @Param        userId path string                  true "User ID"
// @Param        body   body updateMemberRoleRequest true "New role ID"
// @Success      200 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Router       /organizations/{id}/members/{userId}/role [put]
func UpdateMemberRole(c fiber.Ctx) (int, response.APIResponse) {
	orgID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid id"}
	}

	org, err := dal.Organization.GetByID(uint(orgID))
	if err != nil {
		return fiber.StatusNotFound, response.ErrorResponse{Error: "organization not found"}
	}

	if status, errResp := requirePermission(uint(orgID), permissions.MemberRole, c); errResp != nil {
		return status, errResp
	}

	userID := c.Params("userId")
	if userID == "" {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "user id is required"}
	}

	// Prevent changing the owner's role
	if org.OwnerID.String() == userID {
		return fiber.StatusForbidden, response.ErrorResponse{Error: "cannot change owner's role"}
	}

	var req updateMemberRoleRequest
	if err := c.Bind().JSON(&req); err != nil || req.RoleID == 0 {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "roleId is required"}
	}

	// Verify the role belongs to this organization
	role, err := dal.OrgRole.GetByID(req.RoleID)
	if err != nil || role.OrganizationID != uint(orgID) {
		return fiber.StatusNotFound, response.ErrorResponse{Error: "role not found"}
	}

	if err := dal.Organization.UpdateMemberRole(uint(orgID), userID, req.RoleID); err != nil {
		return fiber.StatusInternalServerError, response.ErrorResponse{Error: err.Error()}
	}

	logAudit(c, uint(orgID), nil, models.ActionMemberRole, role.Name, "Changed member '"+userID+"' role to '"+role.Name+"'", "")

	return fiber.StatusOK, response.DataResponse{Data: fiber.Map{"roleId": req.RoleID}}
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

	if controllers.IsInternalOrg(org.Name) && !isInternalUser(c) {
		return fiber.StatusNotFound, response.ErrorResponse{Error: "organization not found"}
	}

	if err := controllers.DeleteOrganization(org); err != nil {
		return fiber.StatusInternalServerError, response.Error500
	}

	logAudit(c, org.ID, nil, models.ActionOrgDelete, org.Name, "Deleted organization '"+org.Name+"'", "")

	return fiber.StatusNoContent, nil
}
