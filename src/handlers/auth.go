package handlers

import (
	"strings"

	"flagpole/src/controllers"
	"flagpole/src/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type signupRequest struct {
	Email     string `json:"email"`
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Password  string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Signup godoc
// @Summary      Register a new user
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        body body signupRequest true "Sign up data"
// @Success      201 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      409 {object} response.ErrorResponse
// @Router       /signup [post]
func Signup(c fiber.Ctx) (int, response.APIResponse) {
	var req signupRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "couldn't parse body"}
	}

	req.Email     = strings.TrimSpace(req.Email)
	req.Username  = strings.TrimSpace(req.Username)
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName  = strings.TrimSpace(req.LastName)

	if req.Email == "" || req.Username == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "email, username, firstName, lastName and password are required"}
	}

	if _, err := controllers.RegisterUser(req.Email, req.Username, req.FirstName, req.LastName, req.Password); err != nil {
		if conflict, ok := err.(*controllers.RegistrationConflict); ok {
			var fields []string
			if conflict.EmailTaken    { fields = append(fields, "email") }
			if conflict.UsernameTaken { fields = append(fields, "username") }
			return fiber.StatusConflict, response.ConflictResponse{Fields: fields}
		}
		return fiber.StatusInternalServerError, response.Error500
	}

	return fiber.StatusCreated, nil
}

// Login godoc
// @Summary      Login and receive a JWT
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        body body loginRequest true "Credentials"
// @Success      200 {object} response.DataResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      401 {object} response.ErrorResponse
// @Router       /login [post]
func Login(c fiber.Ctx) (int, response.APIResponse) {
	var req loginRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "couldn't parse body"}
	}

	req.Email = strings.TrimSpace(req.Email)

	if req.Email == "" || req.Password == "" {
		return fiber.StatusBadRequest, response.ErrorResponse{Error: "email and password are required"}
	}

	user, err := controllers.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		return fiber.StatusUnauthorized, response.ErrInvalidCredentials
	}

	token, err := controllers.GenerateToken(user)
	if err != nil {
		return fiber.StatusInternalServerError, response.Error500
	}

	return fiber.StatusOK, response.DataResponse{Data: fiber.Map{"token": token}}
}
