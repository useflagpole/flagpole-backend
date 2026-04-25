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

var ErrOrgLimitReached = errors.New("organization limit reached")
var ErrInvalidPlan    = errors.New("invalid plan")

const AllowedPlan = "free"

func SetOrganizationPlan(orgID uint, plan string) error {
	if plan != AllowedPlan {
		return ErrInvalidPlan
	}
	return dal.Organization.SetPlan(orgID, plan)
}

const MaxOrgsPerUser = 1

func CreateOrganization(name string, userID uuid.UUID) (*models.Organization, error) {
	owned, err := dal.User.CountOwnedOrganizations(userID)
	if err != nil {
		log.Printf("CreateOrganization: owned org count failed: %v", err)
		return nil, errors.New("internal error")
	}
	if owned >= MaxOrgsPerUser {
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
