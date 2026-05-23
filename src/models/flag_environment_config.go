package models

import (
	"encoding/json"
)

type FlagEnvironmentConfig struct {
	Base
	FlagID            uint        `gorm:"not null;uniqueIndex:idx_flag_env_config"`
	EnvironmentID     uint        `gorm:"not null;uniqueIndex:idx_flag_env_config"`
	Environment       Environment `gorm:"foreignKey:EnvironmentID"                  json:"-"`
	Enabled           bool        `gorm:"not null;default:false"`
	RolloutEnabled    bool        `gorm:"not null;default:false"`
	RolloutPercentage int         `gorm:"not null;default:0"`
	DefaultValue      string      `gorm:"not null"`
	ServedValue       string      `gorm:"not null"`
}

func (FlagEnvironmentConfig) TableName() string {
	return "project.flag_environment_configs"
}

func (c FlagEnvironmentConfig) ParsedDefaultValue() (interface{}, error) {
	var val interface{}
	if err := json.Unmarshal([]byte(c.DefaultValue), &val); err != nil {
		return nil, err
	}
	return val, nil
}

func (c FlagEnvironmentConfig) ParsedServedValue() (interface{}, error) {
	var val interface{}
	if err := json.Unmarshal([]byte(c.ServedValue), &val); err != nil {
		return nil, err
	}
	return val, nil
}
