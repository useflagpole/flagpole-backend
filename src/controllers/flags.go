package controllers

import (
	"encoding/json"
	"errors"

	"flagpole/src/dal"
	"flagpole/src/models"
)

var ErrInvalidFlagName = errors.New("invalid feature flag name")
var ErrFlagAlreadyExists = errors.New("feature flag already exists")
var ErrFlagNotFound = errors.New("feature flag doesn't exist")

func AddFlag(name string, flagType models.FlagType, value interface{}) error {
	if len(name) <= 1 {
		return ErrInvalidFlagName
	}
	if err := models.ValidateValue(flagType, value); err != nil {
		return err
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if err := dal.FeatureFlag.Create(name, string(flagType), string(raw)); err != nil {
		return ErrFlagAlreadyExists
	}
	return nil
}

func GetFlag(name string) (models.FlagValue, error) {
	flag, err := dal.FeatureFlag.FindByName(name)
	if err != nil {
		return models.FlagValue{}, ErrFlagNotFound
	}
	return flag.ToFlagValue()
}

func SetFlag(name string, value interface{}) error {
	flag, err := dal.FeatureFlag.FindByName(name)
	if err != nil {
		return ErrFlagNotFound
	}
	if err := models.ValidateValue(models.FlagType(flag.FlagType), value); err != nil {
		return err
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return dal.FeatureFlag.UpdateValue(&flag, string(raw))
}

func EvaluateFlags(keys []string) map[string]models.FlagValue {
	var flags []models.FeatureFlag
	var err error
	if len(keys) == 0 {
		flags, err = dal.FeatureFlag.FindAll()
	} else {
		flags, err = dal.FeatureFlag.FindByNames(keys)
	}
	if err != nil {
		return map[string]models.FlagValue{}
	}

	result := make(map[string]models.FlagValue)
	for _, f := range flags {
		if fv, err := f.ToFlagValue(); err == nil {
			result[f.Name] = fv
		}
	}
	return result
}
