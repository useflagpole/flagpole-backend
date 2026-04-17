package dal

import (
	"flagpole/src/database"
	"flagpole/src/models"
)

type organizationDAL struct{}

var Organization = organizationDAL{}

func (organizationDAL) List() ([]models.Organization, error) {
	var orgs []models.Organization
	if err := database.DB.Find(&orgs).Error; err != nil {
		return nil, err
	}
	return orgs, nil
}

func (organizationDAL) GetByID(id uint) (*models.Organization, error) {
	var org models.Organization
	if err := database.DB.First(&org, id).Error; err != nil {
		return nil, err
	}
	return &org, nil
}

func (organizationDAL) Create(org *models.Organization) error {
	return database.DB.Create(org).Error
}

func (organizationDAL) Save(org *models.Organization) error {
	return database.DB.Save(org).Error
}

func (organizationDAL) Delete(org *models.Organization) error {
	return database.DB.Delete(org).Error
}
