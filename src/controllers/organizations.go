package controllers

import (
	"errors"
	"log"

	"flagpole/src/dal"
	"flagpole/src/database"
	"flagpole/src/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const INTERNAL_ORG_NAME = "flagpole"
const ALLOWED_PLAN     = "free"
const MAX_ORGS_PER_USER  = 1

var ErrOrgLimitReached  = errors.New("organization limit reached")
var ErrInvalidPlan      = errors.New("invalid plan")
var ErrReservedOrgName  = errors.New("invalid organization name")

func IsInternalOrg(name string) bool {
	return name == INTERNAL_ORG_NAME
}

func IsInternalUser(orgNames []interface{}) bool {
	for _, name := range orgNames {
		if name == INTERNAL_ORG_NAME {
			return true
		}
	}
	return false
}

func SetOrganizationPlan(orgID uint, plan string) error {
	if plan != ALLOWED_PLAN {
		return ErrInvalidPlan
	}
	return dal.Organization.SetPlan(orgID, plan)
}

func CreateOrganization(name string, userID uuid.UUID) (*models.Organization, error) {
	if IsInternalOrg(name) {
		return nil, ErrReservedOrgName
	}

	owned, err := dal.User.CountOwnedOrganizations(userID)
	if err != nil {
		log.Printf("CreateOrganization: owned org count failed: %v", err)
		return nil, errors.New("internal error")
	}
	if owned >= MAX_ORGS_PER_USER {
		return nil, ErrOrgLimitReached
	}

	adminRole, err := dal.Role.GetByName("admin")
	if err != nil {
		log.Printf("CreateOrganization: admin role lookup failed: %v", err)
		return nil, errors.New("internal error")
	}

	var org *models.Organization

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		org = &models.Organization{Name: name, OwnerID: userID}
		if err := tx.Create(org).Error; err != nil {
			return err
		}
		return tx.Create(&models.UserOrganization{
			OrganizationID: org.ID,
			UserID:         userID,
			RoleID:         adminRole.ID,
		}).Error
	})
	if err != nil {
		return nil, err
	}

	return org, nil
}

func UpdateOrganization(org *models.Organization, newName string) (*models.Organization, error) {
	org.Name = newName
	if err := dal.Organization.Save(org); err != nil {
		return nil, err
	}
	return org, nil
}

func DeleteOrganization(org *models.Organization) error {
	return dal.Organization.Delete(org)
}
