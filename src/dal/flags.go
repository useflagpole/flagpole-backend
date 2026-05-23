package dal

import (
	"flagpole/src/database"
	"flagpole/src/models"
)

type featureFlagDAL struct{}

var FeatureFlag = featureFlagDAL{}

func (featureFlagDAL) ListByProject(projectID uint) ([]models.FeatureFlag, error) {
	var flags []models.FeatureFlag
	err := database.DB.
		Select("project.feature_flags.*, COUNT(fec.id) AS env_count").
		Joins("LEFT JOIN project.flag_environment_configs fec ON fec.flag_id = project.feature_flags.id AND fec.deleted_at IS NULL").
		Where("project.feature_flags.project_id = ?", projectID).
		Group("project.feature_flags.id").
		Find(&flags).Error
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

func (featureFlagDAL) UpdateRollout(id uint, rolloutEnabled bool, rolloutPercentage int) error {
	return database.DB.Model(&models.FeatureFlag{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"rollout_enabled":    rolloutEnabled,
			"rollout_percentage": rolloutPercentage,
		}).Error
}

func (featureFlagDAL) UpdateValues(id uint, defaultValue, servedValue string) error {
	return database.DB.Model(&models.FeatureFlag{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"default_value": defaultValue,
			"served_value":  servedValue,
		}).Error
}

