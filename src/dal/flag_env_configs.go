package dal

import (
	"flagpole/src/database"
	"flagpole/src/models"
)

type flagEnvConfigDAL struct{}

var FlagEnvConfig = flagEnvConfigDAL{}

func (flagEnvConfigDAL) GetByFlagAndEnv(flagID uint, env string) (*models.FlagEnvironmentConfig, error) {
	var config models.FlagEnvironmentConfig
	err := database.DB.Where("flag_id = ? AND environment_name = ?", flagID, env).First(&config).Error
	return &config, err
}

func (flagEnvConfigDAL) Create(config *models.FlagEnvironmentConfig) error {
	return database.DB.Create(config).Error
}

func (flagEnvConfigDAL) Save(config *models.FlagEnvironmentConfig) error {
	return database.DB.Save(config).Error
}

func (flagEnvConfigDAL) Delete(config *models.FlagEnvironmentConfig) error {
	return database.DB.Delete(config).Error
}

func (flagEnvConfigDAL) ListByFlag(flagID uint) ([]models.FlagEnvironmentConfig, error) {
	var configs []models.FlagEnvironmentConfig
	err := database.DB.Where("flag_id = ?", flagID).Find(&configs).Error
	return configs, err
}

func (flagEnvConfigDAL) DeleteByEnv(projectID uint, env string) error {
	// Join with feature_flags to get flag IDs for this project
	return database.DB.Exec(`
		DELETE FROM project.flag_environment_configs fec
		USING project.feature_flags ff
		WHERE fec.flag_id = ff.id 
		AND ff.project_id = ? 
		AND fec.environment_name = ?
	`, projectID, env).Error
}

func (flagEnvConfigDAL) Exists(flagID uint, env string) bool {
	var count int64
	database.DB.Model(&models.FlagEnvironmentConfig{}).
		Where("flag_id = ? AND environment_name = ?", flagID, env).
		Count(&count)
	return count > 0
}
