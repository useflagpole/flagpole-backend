package models

import (
	"encoding/json"
)

type Segment struct {
	Base
	ProjectID   uint          `gorm:"not null;uniqueIndex:idx_segment_project_name" json:"projectId"`
	Name        string        `gorm:"not null;uniqueIndex:idx_segment_project_name" json:"name"`
	Description string        `gorm:"default:''"                                    json:"description"`
	UserCount   int           `gorm:"default:0"                                     json:"userCount"`
	Rules       []SegmentRule `gorm:"foreignKey:SegmentID"                          json:"rules,omitempty"`
}

func (Segment) TableName() string {
	return "project.segments"
}

type SegmentRule struct {
	Base
	SegmentID uint   `gorm:"not null" json:"segmentId"`
	Attribute string `gorm:"not null" json:"attribute"`
	Operator  string `gorm:"not null" json:"operator"`
	Value     string `gorm:"not null" json:"value"`
}

func (SegmentRule) TableName() string {
	return "project.segment_rules"
}

type FlagSegmentOverride struct {
	Base
	FlagID    uint    `gorm:"not null;uniqueIndex:idx_override_flag_segment" json:"flagId"`
	SegmentID uint    `gorm:"not null;uniqueIndex:idx_override_flag_segment" json:"segmentId"`
	Value     string  `gorm:"not null"                                       json:"-"`
	Enabled   bool    `gorm:"not null;default:true"                          json:"enabled"`
	Segment   Segment `gorm:"foreignKey:SegmentID"                           json:"segment"`
}

func (FlagSegmentOverride) TableName() string {
	return "project.flag_segment_overrides"
}

func (o FlagSegmentOverride) ParsedValue() (interface{}, error) {
	var val interface{}
	if err := json.Unmarshal([]byte(o.Value), &val); err != nil {
		return nil, err
	}
	return val, nil
}
