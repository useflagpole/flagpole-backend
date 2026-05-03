package dal

import (
	"flagpole/src/database"
	"flagpole/src/models"
)

type flagOverrideDAL struct{}

var FlagOverride = flagOverrideDAL{}

type OverrideWithSegment struct {
	models.FlagSegmentOverride
	SegmentName string `json:"segmentName"`
	UserCount   int    `json:"userCount"`
}

func (flagOverrideDAL) ListByFlag(flagID uint) ([]OverrideWithSegment, error) {
	var overrides []OverrideWithSegment
	err := database.DB.Raw(`
		SELECT fso.id, fso.flag_id, fso.segment_id, fso.value, fso.enabled, fso.created_at, fso.updated_at,
		       s.name AS segment_name, s.user_count
		FROM project.flag_segment_overrides fso
		JOIN project.segments s ON s.id = fso.segment_id
		WHERE fso.flag_id = ?
		ORDER BY s.name ASC
	`, flagID).Scan(&overrides).Error
	return overrides, err
}

func (flagOverrideDAL) GetByFlagAndSegment(flagID, segmentID uint) (*models.FlagSegmentOverride, error) {
	var override models.FlagSegmentOverride
	err := database.DB.Where("flag_id = ? AND segment_id = ?", flagID, segmentID).First(&override).Error
	return &override, err
}

func (flagOverrideDAL) SetOverride(flagID, segmentID uint, value string, enabled bool) error {
	var override models.FlagSegmentOverride
	err := database.DB.Where("flag_id = ? AND segment_id = ?", flagID, segmentID).First(&override).Error
	if err != nil {
		override = models.FlagSegmentOverride{
			FlagID:    flagID,
			SegmentID: segmentID,
			Value:     value,
			Enabled:   enabled,
		}
		return database.DB.Create(&override).Error
	}
	override.Value = value
	override.Enabled = enabled
	return database.DB.Save(&override).Error
}

func (flagOverrideDAL) RemoveOverride(flagID, segmentID uint) error {
	return database.DB.Where("flag_id = ? AND segment_id = ?", flagID, segmentID).Delete(&models.FlagSegmentOverride{}).Error
}
