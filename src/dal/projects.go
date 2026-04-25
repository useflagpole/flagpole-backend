package dal

import (
	"flagpole/src/database"
	"flagpole/src/models"
)

type projectDAL struct{}

var Project = projectDAL{}

func (projectDAL) Create(p *models.Project) error {
	return database.DB.Create(p).Error
}

func (projectDAL) ListByOrg(orgID uint) ([]models.Project, error) {
	var projects []models.Project
	err := database.DB.Where("organization_id = ?", orgID).Find(&projects).Error
	return projects, err
}

func (projectDAL) GetByID(id uint) (*models.Project, error) {
	var p models.Project
	err := database.DB.First(&p, id).Error
	return &p, err
}

func (projectDAL) Save(p *models.Project) error {
	return database.DB.Save(p).Error
}
