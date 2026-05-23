package dal

import (
	"time"

	"flagpole/src/database"
	"flagpole/src/models"

	"gorm.io/gorm"
)

type sdkKeyDAL struct{}

var SDKKey = sdkKeyDAL{}

func (sdkKeyDAL) Create(k *models.SDKKey) error {
	return database.DB.Create(k).Error
}

func (sdkKeyDAL) GetByID(id, projectID uint) (*models.SDKKey, error) {
	var k models.SDKKey
	err := database.DB.Where("id = ? AND project_id = ?", id, projectID).First(&k).Error
	if err != nil {
		return nil, err
	}
	return &k, nil
}

func (sdkKeyDAL) ListByProject(projectID, environmentID uint) ([]models.SDKKey, error) {
	var keys []models.SDKKey
	q := database.DB.Unscoped().Preload("Environment").Where("project_id = ?", projectID)
	if environmentID != 0 {
		q = q.Where("environment_id = ?", environmentID)
	}
	err := q.Order("created_at ASC").Find(&keys).Error
	return keys, err
}

func (sdkKeyDAL) CountActive(projectID, environmentID uint, keyType string) (int64, error) {
	var count int64
	err := database.DB.Model(&models.SDKKey{}).
		Where("project_id = ? AND environment_id = ? AND key_type = ? AND revoked_at IS NULL", projectID, environmentID, keyType).
		Count(&count).Error
	return count, err
}

func (sdkKeyDAL) Revoke(k *models.SDKKey) error {
	now := time.Now()
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(k).Update("revoked_at", now).Error; err != nil {
			return err
		}
		return tx.Delete(k).Error
	})
}
