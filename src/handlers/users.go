package handlers

import (
	"flagpole/src/controllers"
	"flagpole/src/pkg/jwtutil"
	"flagpole/src/pkg/response"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// ListUserOrganizations godoc
// @Summary      List organizations for a user
// @Tags         Users
// @Produce      json
// @Param        user_id path string true "User ID"
// @Success      200 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /users/{user_id}/organizations [get]
func ListUserOrganizations(c fiber.Ctx) (int, response.APIResponse) {
	paramID, err := uuid.Parse(c.Params("user_id"))
	if err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid user id"}
	}

	if paramID != jwtutil.UserID(c) {
		return fiber.StatusForbidden, response.ErrorResponse{Error: "forbidden"}
	}

	orgs, err := controllers.GetUserOrganizations(paramID)
	if err != nil {
		return fiber.StatusInternalServerError, response.Error500
	}

	return fiber.StatusOK, response.DataResponse{Data: orgs}
}
