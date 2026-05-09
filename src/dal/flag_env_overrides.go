package dal

import (
	"flagpole/src/database"
	"flagpole/src/models"
)

type flagEnvOverrideDAL struct{}

var FlagEnvOverride = flagEnvOverrideDAL{}

func (flagEnvOverrideDAL) ListByFlagAndEnv(flagID uint, env string) ([]models.FlagEnvironmentOverride, error) {
	var overrides []models.FlagEnvironmentOverride
	err := database.DB.Where("flag_id = ? AND environment_name = ?", flagID, env).Order("priority ASC").Find(&overrides).Error
	return overrides, err
}

func (flagEnvOverrideDAL) SetOverride(flagID uint, env string, segmentID uint, value string, enabled bool, priority int) error {
	var override models.FlagEnvironmentOverride
	err := database.DB.Where("flag_id = ? AND environment_name = ? AND segment_id = ?", flagID, env, segmentID).First(&override).Error
	if err != nil {
		override = models.FlagEnvironmentOverride{
			FlagID:          flagID,
			EnvironmentName: env,
			SegmentID:       segmentID,
			Value:           value,
			Enabled:         enabled,
			Priority:        priority,
		}
		return database.DB.Create(&override).Error
	}
	override.Value = value
	override.Enabled = enabled
	override.Priority = priority
	return database.DB.Save(&override).Error
}

func (flagEnvOverrideDAL) RemoveOverride(flagID uint, env string, segmentID uint) error {
	return database.DB.Where("flag_id = ? AND environment_name = ? AND segment_id = ?", flagID, env, segmentID).Delete(&models.FlagEnvironmentOverride{}).Error
}

func (flagEnvOverrideDAL) RemoveByEnv(flagID uint, env string) error {
	return database.DB.Where("flag_id = ? AND environment_name = ?", flagID, env).Delete(&models.FlagEnvironmentOverride{}).Error
}

func (flagEnvOverrideDAL) ListFlagsUsingSegment(segmentID uint) ([]models.FlagUsage, error) {
	var results []models.FlagUsage
	err := database.DB.Raw(`
		SELECT DISTINCT ff.id, ff.key
		FROM project.feature_flags ff
		JOIN project.flag_environment_overrides feo ON feo.flag_id = ff.id
		WHERE feo.segment_id = ? AND feo.deleted_at IS NULL AND ff.deleted_at IS NULL
		ORDER BY ff.key ASC
`, segmentID).Scan(&results).Error
	return results, err
}
