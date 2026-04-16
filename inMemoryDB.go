package main

import "errors"

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

func validateValue(flagType FlagType, value interface{}) error {
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

type FeatureFlagMapType map[string]FlagValue

var FeatureFlagMap FeatureFlagMapType

func InitFeatureFlagMap() {
	if FeatureFlagMap == nil {
		FeatureFlagMap = make(FeatureFlagMapType)
	}
}

func (ffmap FeatureFlagMapType) AddFlag(name string, flagType FlagType, value interface{}) error {
	if len(name) <= 1 {
		return errors.New("invalid feature flag name")
	}
	if _, exists := ffmap[name]; exists {
		return errors.New("feature flag already exists")
	}
	if err := validateValue(flagType, value); err != nil {
		return err
	}
	ffmap[name] = FlagValue{Type: flagType, Value: value}
	return nil
}

func (ffmap FeatureFlagMapType) GetFlag(name string) (FlagValue, error) {
	fv, exists := ffmap[name]
	if !exists {
		return FlagValue{}, errors.New("feature flag doesn't exist")
	}
	return fv, nil
}

func (ffmap FeatureFlagMapType) SetFlag(name string, value interface{}) error {
	fv, exists := ffmap[name]
	if !exists {
		return errors.New("feature flag doesn't exist")
	}
	if err := validateValue(fv.Type, value); err != nil {
		return err
	}
	ffmap[name] = FlagValue{Type: fv.Type, Value: value}
	return nil
}
