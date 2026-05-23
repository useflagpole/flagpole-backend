package dal

import (
	"flagpole/src/database"
	"flagpole/src/models"

	"gorm.io/gorm"
)

type flagEnvOverrideDAL struct{}

var FlagEnvOverride = flagEnvOverrideDAL{}

func (f *flagEnvOverrideDAL) db(tx ...*gorm.DB) *gorm.DB {
	if len(tx) > 0 && tx[0] != nil {
		return tx[0]
	}
	return database.DB
}

func (f *flagEnvOverrideDAL) ListByFlagAndEnvID(flagID, envID uint, tx ...*gorm.DB) ([]models.FlagEnvironmentOverride, error) {
	var overrides []models.FlagEnvironmentOverride
	err := f.db(tx...).Where("flag_id = ? AND environment_id = ?", flagID, envID).Order("priority ASC").Find(&overrides).Error
	return overrides, err
}

func (f *flagEnvOverrideDAL) SetOverride(flagID, envID, segmentID uint, value string, enabled bool, priority int, tx ...*gorm.DB) error {
	var override models.FlagEnvironmentOverride
	err := f.db(tx...).Where("flag_id = ? AND environment_id = ? AND segment_id = ?", flagID, envID, segmentID).First(&override).Error
	if err != nil {
		override = models.FlagEnvironmentOverride{
			FlagID:        flagID,
			EnvironmentID: envID,
			SegmentID:     segmentID,
			Value:         value,
			Enabled:       enabled,
			Priority:      priority,
		}
		return f.db(tx...).Create(&override).Error
	}
	override.Value = value
	override.Enabled = enabled
	override.Priority = priority
	return f.db(tx...).Save(&override).Error
}

func (f *flagEnvOverrideDAL) RemoveOverride(flagID, envID, segmentID uint, tx ...*gorm.DB) error {
	return f.db(tx...).Where("flag_id = ? AND environment_id = ? AND segment_id = ?", flagID, envID, segmentID).Delete(&models.FlagEnvironmentOverride{}).Error
}

func (f *flagEnvOverrideDAL) RemoveByEnvID(flagID, envID uint, tx ...*gorm.DB) error {
	return f.db(tx...).Where("flag_id = ? AND environment_id = ?", flagID, envID).Delete(&models.FlagEnvironmentOverride{}).Error
}

func (f *flagEnvOverrideDAL) ListFlagsUsingSegment(segmentID uint, tx ...*gorm.DB) ([]models.FlagUsage, error) {
	var results []models.FlagUsage
	err := f.db(tx...).Raw(`
		SELECT DISTINCT ff.id, ff.key
		FROM project.feature_flags ff
		JOIN project.flag_environment_overrides feo ON feo.flag_id = ff.id
		WHERE feo.segment_id = ? AND feo.deleted_at IS NULL AND ff.deleted_at IS NULL
		ORDER BY ff.key ASC
	`, segmentID).Scan(&results).Error
	return results, err
}
