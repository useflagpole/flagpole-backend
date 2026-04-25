package models

import (
	"encoding/json"
	"errors"

	"gorm.io/gorm"
)

type FlagType string

const (
	FLAG_TYPE_BOOL   FlagType = "bool"
	FLAG_TYPE_STRING FlagType = "string"
	FLAG_TYPE_NUMBER FlagType = "number"
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
	case FLAG_TYPE_BOOL:
		if _, ok := value.(bool); !ok {
			return errors.New("value must be a boolean")
		}
	case FLAG_TYPE_STRING:
		if _, ok := value.(string); !ok {
			return errors.New("value must be a string")
		}
	case FLAG_TYPE_NUMBER:
		if _, ok := value.(float64); !ok {
			return errors.New("value must be a number")
		}
	default:
		return errors.New("invalid flag type, must be bool, string, or number")
	}
	return nil
}
