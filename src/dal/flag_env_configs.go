package dal

import (
	"flagpole/src/database"
	"flagpole/src/models"

	"gorm.io/gorm"
)

type flagEnvConfigDAL struct{}

var FlagEnvConfig = flagEnvConfigDAL{}

func (f *flagEnvConfigDAL) db(tx ...*gorm.DB) *gorm.DB {
	if len(tx) > 0 && tx[0] != nil {
		return tx[0]
	}
	return database.DB
}

func (f *flagEnvConfigDAL) GetByFlagAndEnvID(flagID, envID uint, tx ...*gorm.DB) (*models.FlagEnvironmentConfig, error) {
	var config models.FlagEnvironmentConfig
	err := f.db(tx...).Preload("Environment").Where("flag_id = ? AND environment_id = ?", flagID, envID).First(&config).Error
	return &config, err
}

func (f *flagEnvConfigDAL) Create(config *models.FlagEnvironmentConfig) error {
	return database.DB.Create(config).Error
}

func (f *flagEnvConfigDAL) Save(config *models.FlagEnvironmentConfig, tx ...*gorm.DB) error {
	return f.db(tx...).Save(config).Error
}

func (f *flagEnvConfigDAL) Delete(config *models.FlagEnvironmentConfig) error {
	return database.DB.Delete(config).Error
}

func (f *flagEnvConfigDAL) ListByFlag(flagID uint) ([]models.FlagEnvironmentConfig, error) {
	var configs []models.FlagEnvironmentConfig
	err := database.DB.Where("flag_id = ?", flagID).Find(&configs).Error
	return configs, err
}

func (f *flagEnvConfigDAL) DeleteByEnvID(projectID, envID uint, tx ...*gorm.DB) error {
	return f.db(tx...).Exec(`
		DELETE FROM project.flag_environment_configs fec
		USING project.feature_flags ff
		WHERE fec.flag_id = ff.id
		AND ff.project_id = ?
		AND fec.environment_id = ?
	`, projectID, envID).Error
}

func (f *flagEnvConfigDAL) Exists(flagID, envID uint, tx ...*gorm.DB) bool {
	var count int64
	f.db(tx...).Model(&models.FlagEnvironmentConfig{}).
		Where("flag_id = ? AND environment_id = ?", flagID, envID).
		Count(&count)
	return count > 0
}
