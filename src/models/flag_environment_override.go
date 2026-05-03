package models

import (
	"encoding/json"
)

type FlagEnvironmentOverride struct {
	Base
	FlagID          uint   `gorm:"not null;uniqueIndex:idx_flag_env_seg"`
	EnvironmentName string `gorm:"not null;uniqueIndex:idx_flag_env_seg"`
	SegmentID       uint   `gorm:"not null;uniqueIndex:idx_flag_env_seg"`
	Value           string `gorm:"not null"`
	Enabled         bool   `gorm:"not null;default:true"`
}

func (FlagEnvironmentOverride) TableName() string {
	return "project.flag_environment_overrides"
}

func (o FlagEnvironmentOverride) ParsedValue() (interface{}, error) {
	var val interface{}
	if err := json.Unmarshal([]byte(o.Value), &val); err != nil {
		return nil, err
	}
	return val, nil
}
