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

func (projectDAL) ListByOrg(orgID uint, includeArchived bool) ([]models.Project, error) {
	var projects []models.Project
	q := database.DB.Where("organization_id = ?", orgID)
	if !includeArchived {
		q = q.Where("is_active = true")
	}
	err := q.Find(&projects).Error
	return projects, err
}

func (projectDAL) GetByID(id uint) (*models.Project, error) {
	var p models.Project
	err := database.DB.First(&p, id).Error
	return &p, err
}

func (projectDAL) CountByOrg(orgID uint) (int64, error) {
	var count int64
	err := database.DB.Model(&models.Project{}).Where("organization_id = ? AND is_active = true", orgID).Count(&count).Error
	return count, err
}

func (projectDAL) Save(p *models.Project) error {
	return database.DB.Save(p).Error
}
