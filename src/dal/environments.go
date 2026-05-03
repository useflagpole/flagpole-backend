package dal

import (
	"flagpole/src/database"
	"flagpole/src/models"
)

type environmentDAL struct{}

var Environment = environmentDAL{}

func (environmentDAL) ListByProject(projectID uint) ([]models.Environment, error) {
	var envs []models.Environment
	err := database.DB.Where("project_id = ?", projectID).Order("id ASC").Find(&envs).Error
	return envs, err
}

func (environmentDAL) GetByName(projectID uint, name string) (*models.Environment, error) {
	var env models.Environment
	err := database.DB.Where("project_id = ? AND name = ?", projectID, name).First(&env).Error
	return &env, err
}

func (environmentDAL) Create(env *models.Environment) error {
	return database.DB.Create(env).Error
}

func (environmentDAL) Save(env *models.Environment) error {
	return database.DB.Save(env).Error
}

func (environmentDAL) Delete(env *models.Environment) error {
	return database.DB.Delete(env).Error
}

func (environmentDAL) NameExists(projectID uint, name string) bool {
	var count int64
	database.DB.Model(&models.Environment{}).
		Where("project_id = ? AND name = ?", projectID, name).
		Count(&count)
	return count > 0
}

func (environmentDAL) MigrateFromProject(projectID uint, names []string) error {
	for _, name := range names {
		env := models.Environment{
			ProjectID: projectID,
			Name:      name,
		}
		if err := database.DB.FirstOrCreate(&env, models.Environment{ProjectID: projectID, Name: name}).Error; err != nil {
			return err
		}
	}
	return nil
}
