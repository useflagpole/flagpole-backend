package handlers

import (
	"log"
	"strconv"

	"flagpole/src/dal"
	"flagpole/src/models"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

const internalOrgName = "Flagpole"

type orgRequest struct {
	Name string `json:"name"`
}

type orgResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func isInternalUser(c fiber.Ctx) bool {
	claims, ok := c.Locals("claims").(jwt.MapClaims)
	if !ok {
		return false
	}
	orgName, _ := claims["orgName"].(string)
	return orgName == internalOrgName
}

// ListOrganizations godoc
// @Summary      List all organizations
// @Tags         Organizations
// @Produce      json
// @Success      200 {array}  orgResponse
// @Failure      500 {string} string "internal error"
// @Router       /organizations [get]
func ListOrganizations(c fiber.Ctx) error {
	orgs, err := dal.Organization.List()
	if err != nil {
		log.Printf("ListOrganizations: %v", err)
		return c.SendStatus(500)
	}

	if !isInternalUser(c) {
		filtered := make([]models.Organization, 0, len(orgs))
		for _, o := range orgs {
			if o.Name != internalOrgName {
				filtered = append(filtered, o)
			}
		}
		return c.JSON(filtered)
	}

	return c.JSON(orgs)
}

// GetOrganization godoc
// @Summary      Get an organization by ID
// @Tags         Organizations
// @Produce      json
// @Param        id path int true "Organization ID"
// @Success      200 {object} orgResponse
// @Failure      400 {string} string "invalid id"
// @Failure      404 {string} string "not found"
// @Router       /organizations/{id} [get]
func GetOrganization(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("invalid id")
	}

	org, err := dal.Organization.GetByID(uint(id))
	if err != nil {
		return c.Status(404).SendString("organization not found")
	}

	if org.Name == internalOrgName && !isInternalUser(c) {
		return c.Status(404).SendString("organization not found")
	}

	return c.JSON(org)
}

// CreateOrganization godoc
// @Summary      Create an organization
// @Tags         Organizations
// @Accept       json
// @Produce      json
// @Param        body body orgRequest true "Organization data"
// @Success      201 {object} orgResponse
// @Failure      400 {string} string "bad request"
// @Router       /organizations [post]
func CreateOrganization(c fiber.Ctx) error {
	var req orgRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).SendString("couldn't parse body")
	}
	if req.Name == "" {
		return c.Status(400).SendString("name is required")
	}
	if req.Name == internalOrgName {
		return c.Status(400).SendString("invalid organization name")
	}

	org := &models.Organization{Name: req.Name}
	if err := dal.Organization.Create(org); err != nil {
		log.Printf("CreateOrganization: %v", err)
		return c.SendStatus(500)
	}

	return c.Status(201).JSON(org)
}

// UpdateOrganization godoc
// @Summary      Update an organization
// @Tags         Organizations
// @Accept       json
// @Produce      json
// @Param        id   path int        true "Organization ID"
// @Param        body body orgRequest true "Organization data"
// @Success      200 {object} orgResponse
// @Failure      400 {string} string "bad request"
// @Failure      404 {string} string "not found"
// @Router       /organizations/{id} [put]
func UpdateOrganization(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("invalid id")
	}

	org, err := dal.Organization.GetByID(uint(id))
	if err != nil {
		return c.Status(404).SendString("organization not found")
	}

	if org.Name == internalOrgName && !isInternalUser(c) {
		return c.Status(404).SendString("organization not found")
	}

	var req orgRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).SendString("couldn't parse body")
	}
	if req.Name == "" {
		return c.Status(400).SendString("name is required")
	}

	org.Name = req.Name
	if err := dal.Organization.Save(org); err != nil {
		log.Printf("UpdateOrganization: %v", err)
		return c.SendStatus(500)
	}

	return c.JSON(org)
}

// DeleteOrganization godoc
// @Summary      Delete an organization
// @Tags         Organizations
// @Param        id path int true "Organization ID"
// @Success      204
// @Failure      400 {string} string "invalid id"
// @Failure      404 {string} string "not found"
// @Router       /organizations/{id} [delete]
func DeleteOrganization(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("invalid id")
	}

	org, err := dal.Organization.GetByID(uint(id))
	if err != nil {
		return c.Status(404).SendString("organization not found")
	}

	if org.Name == internalOrgName && !isInternalUser(c) {
		return c.Status(404).SendString("organization not found")
	}

	if err := dal.Organization.Delete(org); err != nil {
		log.Printf("DeleteOrganization: %v", err)
		return c.SendStatus(500)
	}

	return c.SendStatus(204)
}
