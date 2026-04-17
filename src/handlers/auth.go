package handlers

import (
	"log"
	"strings"
	"time"

	"flagpole/src/config"
	"flagpole/src/controllers"
	"flagpole/src/dal"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

type signupRequest struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Password  string `json:"password"`
	OrgID     uint   `json:"orgId"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Signup godoc
// @Summary      Register a new user
// @Tags         Authentication
// @Accept       json
// @Produce      plain
// @Param        body body signupRequest true "Sign up data"
// @Success      201
// @Failure      400 {string} string "bad request"
// @Failure      409 {string} string "email already in use"
// @Router       /signup [post]
func Signup(c fiber.Ctx) error {
	var req signupRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).SendString("couldn't parse body")
	}

	req.Email = strings.TrimSpace(req.Email)
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)

	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		return c.Status(400).SendString("email, firstName, lastName and password are required")
	}
	if req.OrgID == 0 {
		return c.Status(400).SendString("orgId is required")
	}

	if _, err := controllers.RegisterUser(req.Email, req.FirstName, req.LastName, req.Password, req.OrgID); err != nil {
		if err == controllers.ErrEmailAlreadyRegistered {
			return c.Status(409).SendString(err.Error())
		}
		return c.SendStatus(500)
	}

	return c.SendStatus(201)
}

// Login godoc
// @Summary      Login and receive a JWT
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        body body loginRequest true "Credentials"
// @Success      200 {object} map[string]string
// @Failure      400 {string} string "bad request"
// @Failure      401 {string} string "invalid credentials"
// @Router       /login [post]
func Login(c fiber.Ctx) error {
	var req loginRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).SendString("couldn't parse body")
	}

	req.Email = strings.TrimSpace(req.Email)

	if req.Email == "" || req.Password == "" {
		return c.Status(400).SendString("email and password are required")
	}

	user, err := controllers.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		return c.Status(401).SendString(err.Error())
	}

	role, err := dal.Role.GetByID(user.RoleID)
	if err != nil {
		log.Printf("Login: role lookup failed for user %s: %v", user.ID, err)
		return c.SendStatus(500)
	}

	org, err := dal.Organization.GetByID(user.OrgID)
	if err != nil {
		log.Printf("Login: org lookup failed for user %s: %v", user.ID, err)
		return c.SendStatus(500)
	}

	claims := jwt.MapClaims{
		"userId":    user.ID,
		"firstName": user.FirstName,
		"lastName":  user.LastName,
		"email":     user.Email,
		"role":      role.Name,
		"orgId":     user.OrgID,
		"orgName":   org.Name,
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(config.Get().JWTSecret))
	if err != nil {
		return c.SendStatus(500)
	}

	return c.JSON(fiber.Map{"token": token})
}
