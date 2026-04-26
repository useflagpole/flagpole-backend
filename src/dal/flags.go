package dal

import (
	"flagpole/src/database"
	"flagpole/src/models"
)

type featureFlagDAL struct{}

var FeatureFlag = featureFlagDAL{}

func (featureFlagDAL) ListByProject(projectID uint) ([]models.FeatureFlag, error) {
	var flags []models.FeatureFlag
	err := database.DB.Where("project_id = ?", projectID).Find(&flags).Error
	return flags, err
}

func (featureFlagDAL) GetByID(id uint, projectID uint) (*models.FeatureFlag, error) {
	var flag models.FeatureFlag
	err := database.DB.Where("id = ? AND project_id = ?", id, projectID).First(&flag).Error
	if err != nil {
		return nil, err
	}
	return &flag, nil
}

func (featureFlagDAL) CountByProject(projectID uint) (int64, error) {
	var count int64
	err := database.DB.Model(&models.FeatureFlag{}).Where("project_id = ?", projectID).Count(&count).Error
	return count, err
}

func (featureFlagDAL) KeyExists(projectID uint, key string) bool {
	var count int64
	database.DB.Model(&models.FeatureFlag{}).Where("project_id = ? AND key = ?", projectID, key).Count(&count)
	return count > 0
}

func (featureFlagDAL) Create(flag *models.FeatureFlag) error {
	return database.DB.Create(flag).Error
}

func (featureFlagDAL) Save(flag *models.FeatureFlag) error {
	return database.DB.Save(flag).Error
}

func (featureFlagDAL) Delete(flag *models.FeatureFlag) error {
	return database.DB.Delete(flag).Error
}
