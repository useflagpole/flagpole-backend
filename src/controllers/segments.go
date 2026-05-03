package controllers

import (
	"errors"

	"flagpole/src/dal"
	"flagpole/src/models"
)

var (
	ErrSegmentNameTaken = errors.New("a segment with that name already exists in this project")
	ErrSegmentNotFound  = errors.New("segment not found")
)

func ListSegments(projectID uint) ([]models.Segment, error) {
	return dal.Segment.ListByProject(projectID)
}

func GetSegment(projectID uint, segmentID uint) (*models.Segment, error) {
	seg, err := dal.Segment.GetByID(segmentID, projectID)
	if err != nil {
		return nil, ErrSegmentNotFound
	}
	rules, err := dal.Segment.GetRules(seg.ID)
	if err != nil {
		return nil, err
	}
	seg.Rules = rules
	return seg, nil
}

func CreateSegment(projectID uint, name, description string, rules []models.SegmentRule) (*models.Segment, error) {
	if dal.Segment.NameExists(projectID, name) {
		return nil, ErrSegmentNameTaken
	}
	seg := &models.Segment{
		ProjectID:   projectID,
		Name:        name,
		Description: description,
	}
	if err := dal.Segment.Create(seg); err != nil {
		return nil, err
	}
	if len(rules) > 0 {
		for i := range rules {
			rules[i].SegmentID = seg.ID
		}
		if err := dal.Segment.SetRules(seg.ID, rules); err != nil {
			return nil, err
		}
	}
	return seg, nil
}

func UpdateSegment(segment *models.Segment, name, description string, rules []models.SegmentRule) (*models.Segment, error) {
	if name != "" && name != segment.Name {
		if dal.Segment.NameExists(segment.ProjectID, name) {
			return nil, ErrSegmentNameTaken
		}
		segment.Name = name
	}
	if description != "" {
		segment.Description = description
	}
	if err := dal.Segment.Save(segment); err != nil {
		return nil, err
	}
	if rules != nil {
		for i := range rules {
			rules[i].SegmentID = segment.ID
		}
		if err := dal.Segment.SetRules(segment.ID, rules); err != nil {
			return nil, err
		}
	}
	rules, err := dal.Segment.GetRules(segment.ID)
	if err != nil {
		return nil, err
	}
	segment.Rules = rules
	return segment, nil
}

func DeleteSegment(segment *models.Segment) error {
	return dal.Segment.Delete(segment)
}
