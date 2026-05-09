package dal

import (
	"flagpole/src/database"
	"flagpole/src/models"
)

type segmentDAL struct{}

var Segment = segmentDAL{}

func (segmentDAL) ListByProject(projectID uint) ([]models.Segment, error) {
	var segments []models.Segment
	err := database.DB.
		Select("project.segments.*, (SELECT COUNT(*) FROM project.segment_rules sr WHERE sr.segment_id = project.segments.id AND sr.deleted_at IS NULL) as rule_count").
		Where("project_id = ?", projectID).
		Order("name ASC").
		Find(&segments).Error
	return segments, err
}

func (segmentDAL) GetByID(id uint, projectID uint) (*models.Segment, error) {
	var seg models.Segment
	err := database.DB.Where("id = ? AND project_id = ?", id, projectID).First(&seg).Error
	if err != nil {
		return nil, err
	}
	return &seg, nil
}

func (segmentDAL) GetByName(projectID uint, name string) (*models.Segment, error) {
	var seg models.Segment
	err := database.DB.Where("project_id = ? AND name = ?", projectID, name).First(&seg).Error
	return &seg, err
}

func (segmentDAL) Create(seg *models.Segment) error {
	return database.DB.Create(seg).Error
}

func (segmentDAL) Save(seg *models.Segment) error {
	return database.DB.Save(seg).Error
}

func (segmentDAL) Delete(seg *models.Segment) error {
	return database.DB.Delete(seg).Error
}

func (segmentDAL) NameExists(projectID uint, name string) bool {
	var count int64
	database.DB.Model(&models.Segment{}).Where("project_id = ? AND name = ?", projectID, name).Count(&count)
	return count > 0
}

func (segmentDAL) GetRules(segmentID uint) ([]models.SegmentRule, error) {
	var rules []models.SegmentRule
	err := database.DB.Where("segment_id = ?", segmentID).Order("id ASC").Find(&rules).Error
	return rules, err
}

func (segmentDAL) SetRules(segmentID uint, rules []models.SegmentRule) error {
	if err := database.DB.Where("segment_id = ?", segmentID).Delete(&models.SegmentRule{}).Error; err != nil {
		return err
	}
	if len(rules) > 0 {
		return database.DB.Create(&rules).Error
	}
	return nil
}
