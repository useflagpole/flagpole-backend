package handlers

import (
	"strings"

	"flagpole/src/controllers"
	"flagpole/src/pkg/jwtutil"
	"flagpole/src/pkg/response"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// GetUser godoc
// @Summary      Get current user profile
// @Tags         Users
// @Produce      json
// @Param        user_id path string true "User ID"
// @Success      200 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /users/{user_id} [get]
func GetUser(c fiber.Ctx) (int, response.APIResponse) {
	paramID, err := uuid.Parse(c.Params("user_id"))
	if err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid user id"}
	}
	if paramID != jwtutil.UserID(c) {
		return fiber.StatusForbidden, response.ErrorResponse{Error: "forbidden"}
	}

	dto, err := controllers.GetUser(paramID)
	if err != nil {
		return fiber.StatusInternalServerError, response.Error500
	}

	return fiber.StatusOK, response.DataResponse{Data: dto}
}

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

// UpdateUsername godoc
// @Summary      Update username for a user
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        user_id path string true "User ID"
// @Param        body body object true "Username payload"
// @Success      200 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      409 {object} response.ConflictResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /users/{user_id}/username [patch]
func UpdateUsername(c fiber.Ctx) (int, response.APIResponse) {
	paramID, err := uuid.Parse(c.Params("user_id"))
	if err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "invalid user id"}
	}
	if paramID != jwtutil.UserID(c) {
		return fiber.StatusForbidden, response.ErrorResponse{Error: "forbidden"}
	}

	var body struct {
		Username string `json:"username"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "couldn't parse body"}
	}
	body.Username = strings.TrimSpace(body.Username)
	if body.Username == "" {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "username is required"}
	}

	if err := controllers.UpdateUsername(paramID, body.Username); err != nil {
		if err == controllers.ErrUsernameTaken {
			return fiber.StatusConflict, response.ConflictResponse{Fields: []string{"username"}}
		}
		return fiber.StatusInternalServerError, response.Error500
	}

	return fiber.StatusOK, response.DataResponse{Data: fiber.Map{"username": body.Username}}
}
