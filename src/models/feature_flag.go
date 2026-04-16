package models

import (
	"encoding/json"
	"errors"

	"gorm.io/gorm"
)

type FlagType string

const (
	FlagTypeBool   FlagType = "bool"
	FlagTypeString FlagType = "string"
	FlagTypeNumber FlagType = "number"
)

type FlagValue struct {
	Type  FlagType    `json:"type"`
	Value interface{} `json:"value"`
}

type FeatureFlag struct {
	gorm.Model
	Name     string `gorm:"uniqueIndex;not null"`
	FlagType string `gorm:"not null"`
	RawValue string `gorm:"not null"`
}

func (f FeatureFlag) ToFlagValue() (FlagValue, error) {
	var val interface{}
	if err := json.Unmarshal([]byte(f.RawValue), &val); err != nil {
		return FlagValue{}, err
	}
	return FlagValue{Type: FlagType(f.FlagType), Value: val}, nil
}

func ValidateValue(flagType FlagType, value interface{}) error {
	switch flagType {
	case FlagTypeBool:
		if _, ok := value.(bool); !ok {
			return errors.New("value must be a boolean")
		}
	case FlagTypeString:
		if _, ok := value.(string); !ok {
			return errors.New("value must be a string")
		}
	case FlagTypeNumber:
		if _, ok := value.(float64); !ok {
			return errors.New("value must be a number")
		}
	default:
		return errors.New("invalid flag type, must be bool, string, or number")
	}
	return nil
}
