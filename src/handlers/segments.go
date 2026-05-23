package handlers

import (
	"errors"
	"log"
	"strconv"

	"flagpole/src/controllers"
	"flagpole/src/dal"
	"flagpole/src/models"
	"flagpole/src/pkg/permissions"
	"flagpole/src/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type segmentCreateRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	MatchType   string                 `json:"matchType"`
	Rules       []models.SegmentRule   `json:"rules"`
}

type segmentUpdateRequest struct {
	Name        *string               `json:"name"`
	Description *string               `json:"description"`
	MatchType   *string               `json:"matchType"`
	Rules       *[]models.SegmentRule `json:"rules"`
}

func segmentErr(err error) (int, response.APIResponse) {
	switch {
	case errors.Is(err, controllers.ErrSegmentNameTaken):
		return fiber.StatusConflict, response.ErrorResponse{Error: err.Error()}
	case errors.Is(err, controllers.ErrSegmentNotFound):
		return fiber.StatusNotFound, response.ErrorResponse{Error: err.Error()}
	default:
		log.Printf("segmentErr: %v", err)
		return fiber.StatusInternalServerError, response.Error500
	}
}

func resolveSegment(c fiber.Ctx) (*models.Segment, *models.Project, int, response.APIResponse) {
	proj, status, errResp := resolveProject(c)
	if errResp != nil {
		return nil, nil, status, errResp
	}
	segmentID, err := strconv.ParseUint(c.Params("segment_id"), 10, 64)
	if err != nil {
		return nil, nil, fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid segment id"}
	}
	seg, err := controllers.GetSegment(proj.ID, uint(segmentID))
	if err != nil {
		return nil, nil, fiber.StatusNotFound, response.ErrorResponse{Error: err.Error()}
	}
	return seg, proj, 0, nil
}

// ListSegments godoc
// @Summary      List segments for a project
// @Tags         Segments
// @Produce      json
// @Param        org_id     path int true "Organization ID"
// @Param        project_id path int true "Project ID"
// @Success      200 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Router       /organizations/{org_id}/projects/{project_id}/segments [get]
func ListSegments(c fiber.Ctx) (int, response.APIResponse) {
	proj, status, errResp := resolveProject(c)
	if errResp != nil {
		return status, errResp
	}
	segments, err := controllers.ListSegments(proj.ID)
	if err != nil {
		return fiber.StatusInternalServerError, response.Error500
	}
	return fiber.StatusOK, response.DataResponse{Data: segments}
}

// CreateSegment godoc
// @Summary      Create a segment
// @Tags         Segments
// @Accept       json
// @Produce      json
// @Param        org_id     path int true "Organization ID"
// @Param        project_id path int true "Project ID"
// @Param        body       body segmentCreateRequest true "Segment definition"
// @Success      201 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      409 {object} response.ErrorResponse
// @Router       /organizations/{org_id}/projects/{project_id}/segments [post]
func CreateSegment(c fiber.Ctx) (int, response.APIResponse) {
	proj, status, errResp := resolveProject(c)
	if errResp != nil {
		return status, errResp
	}
	var req segmentCreateRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "couldn't parse body"}
	}
	if status, errResp := requirePermission(proj.OrganizationID, permissions.SegmentCreate, c); errResp != nil {
		return status, errResp
	}
	seg, err := controllers.CreateSegment(proj.ID, req.Name, req.Description, req.MatchType, req.Rules)
	if err != nil {
		return segmentErr(err)
	}
	logAudit(c, proj.OrganizationID, &proj.ID, models.ActionSegmentCreate, seg.Name, "Created segment '"+seg.Name+"'", "")
	return fiber.StatusCreated, response.DataResponse{Data: seg}
}

// GetSegment godoc
// @Summary      Get a segment by ID
// @Tags         Segments
// @Produce      json
// @Param        org_id     path int true "Organization ID"
// @Param        project_id path int true "Project ID"
// @Param        segment_id path int true "Segment ID"
// @Success      200 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Router       /organizations/{org_id}/projects/{project_id}/segments/{segment_id} [get]
func GetSegment(c fiber.Ctx) (int, response.APIResponse) {
	seg, proj, status, errResp := resolveSegment(c)
	if errResp != nil {
		return status, errResp
	}
	segment, err := controllers.GetSegment(proj.ID, seg.ID)
	if err != nil {
		return segmentErr(err)
	}
	flags, err := dal.FlagEnvOverride.ListFlagsUsingSegment(seg.ID)
	if err != nil {
		return fiber.StatusInternalServerError, response.Error500
	}
	segment.FlagsUsing = flags
	return fiber.StatusOK, response.DataResponse{Data: segment}
}

// UpdateSegment godoc
// @Summary      Update a segment
// @Tags         Segments
// @Accept       json
// @Produce      json
// @Param        org_id     path int true "Organization ID"
// @Param        project_id path int true "Project ID"
// @Param        segment_id path int true "Segment ID"
// @Param        body       body segmentUpdateRequest true "Fields to update"
// @Success      200 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      409 {object} response.ErrorResponse
// @Router       /organizations/{org_id}/projects/{project_id}/segments/{segment_id} [patch]
func UpdateSegment(c fiber.Ctx) (int, response.APIResponse) {
	seg, proj, status, errResp := resolveSegment(c)
	if errResp != nil {
		return status, errResp
	}
	var req segmentUpdateRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "couldn't parse body"}
	}
	if status, errResp := requirePermission(proj.OrganizationID, permissions.SegmentEdit, c); errResp != nil {
		return status, errResp
	}
	name := ""
	if req.Name != nil {
		name = *req.Name
	}
	desc := ""
	if req.Description != nil {
		desc = *req.Description
	}
	matchType := ""
	if req.MatchType != nil {
		matchType = *req.MatchType
	}
	var rules []models.SegmentRule
	if req.Rules != nil {
		rules = *req.Rules
	}
	updated, err := controllers.UpdateSegment(seg, name, desc, matchType, rules)
	if err != nil {
		return segmentErr(err)
	}
	logAudit(c, proj.OrganizationID, &proj.ID, models.ActionSegmentUpdate, updated.Name, "Updated segment '"+updated.Name+"'", "")
	return fiber.StatusOK, response.DataResponse{Data: updated}
}

// DeleteSegment godoc
// @Summary      Delete a segment
// @Tags         Segments
// @Produce      json
// @Param        org_id     path int true "Organization ID"
// @Param        project_id path int true "Project ID"
// @Param        segment_id path int true "Segment ID"
// @Success      204
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /organizations/{org_id}/projects/{project_id}/segments/{segment_id} [delete]
func DeleteSegment(c fiber.Ctx) (int, response.APIResponse) {
	seg, proj, status, errResp := resolveSegment(c)
	if errResp != nil {
		return status, errResp
	}
	if status, errResp := requirePermission(proj.OrganizationID, permissions.SegmentDelete, c); errResp != nil {
		return status, errResp
	}
	if err := controllers.DeleteSegment(seg); err != nil {
		return fiber.StatusInternalServerError, response.Error500
	}
	logAudit(c, proj.OrganizationID, &proj.ID, models.ActionSegmentDelete, seg.Name, "Deleted segment '"+seg.Name+"'", "")
	return fiber.StatusNoContent, nil
}
