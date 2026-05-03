package models

import (
	"errors"
	"fmt"
)

type FlagType string

const (
	FLAG_TYPE_BOOL   FlagType = "bool"
	FLAG_TYPE_STRING FlagType = "string"
	FLAG_TYPE_NUMBER FlagType = "number"
)

const FLAG_STRING_MAX_LEN = 50

type FeatureFlag struct {
	Base
	ProjectID   uint   `gorm:"not null;uniqueIndex:idx_flag_key_project"   json:"projectId"`
	Key         string `gorm:"not null;uniqueIndex:idx_flag_key_project"   json:"key"`
	Description string `gorm:"not null;default:''"                         json:"description"`
	FlagType    string `gorm:"not null"                                    json:"type"`
}

func (FeatureFlag) TableName() string {
	return "project.feature_flags"
}

type FlagValue struct {
	Type  FlagType    `json:"type"`
	Value interface{} `json:"value"`
}

func ValidateValue(flagType FlagType, value interface{}) error {
	switch flagType {
	case FLAG_TYPE_BOOL:
		if _, ok := value.(bool); !ok {
			return errors.New("value must be a boolean")
		}
	case FLAG_TYPE_STRING:
		s, ok := value.(string)
		if !ok {
			return errors.New("value must be a string")
		}
		if len(s) > FLAG_STRING_MAX_LEN {
			return errors.New("string value exceeds " + fmt.Sprint(FLAG_STRING_MAX_LEN) + " characters")
		}
	case FLAG_TYPE_NUMBER:
		if _, ok := value.(float64); !ok {
			return errors.New("value must be a number")
		}
	default:
		return errors.New("invalid flag type: must be bool, string, or number")
	}
	return nil
}
