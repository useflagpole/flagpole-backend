package dal

import (
	"flagpole/src/database"
	"flagpole/src/models"
)

type featureFlagDAL struct{}

var FeatureFlag = featureFlagDAL{}

func (featureFlagDAL) Create(name, flagType, rawValue string) error {
	return database.DB.Create(&models.FeatureFlag{Name: name, FlagType: flagType, RawValue: rawValue}).Error
}

func (featureFlagDAL) FindByName(name string) (models.FeatureFlag, error) {
	var flag models.FeatureFlag
	err := database.DB.Where("name = ?", name).First(&flag).Error
	return flag, err
}

func (featureFlagDAL) UpdateValue(flag *models.FeatureFlag, rawValue string) error {
	return database.DB.Model(flag).Update("raw_value", rawValue).Error
}

func (featureFlagDAL) FindAll() ([]models.FeatureFlag, error) {
	var flags []models.FeatureFlag
	err := database.DB.Find(&flags).Error
	return flags, err
}

func (featureFlagDAL) FindByNames(names []string) ([]models.FeatureFlag, error) {
	var flags []models.FeatureFlag
	err := database.DB.Where("name IN ?", names).Find(&flags).Error
	return flags, err
}
