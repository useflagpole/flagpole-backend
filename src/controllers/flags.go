package controllers

import (
	"encoding/json"
	"errors"
	"regexp"

	"flagpole/src/dal"
	"flagpole/src/models"
)

const MAX_FLAGS_PER_PROJECT = 25
const FLAG_KEY_MAX_LEN      = 64

var (
	ErrFlagKeyInvalid   = errors.New("key must be 2–64 chars, lowercase letters, numbers, hyphens or underscores")
	ErrFlagKeyTaken     = errors.New("a flag with that key already exists in this project")
	ErrFlagNotFound     = errors.New("flag not found")
	ErrFlagLimitReached = errors.New("project has reached the maximum of 25 flags")
)

var flagKeyRe = regexp.MustCompile(`^[a-z0-9_-]{2,64}$`) // min 2, max FLAG_KEY_MAX_LEN

func ListFlags(projectID uint) ([]models.FeatureFlag, error) {
	return dal.FeatureFlag.ListByProject(projectID)
}

func CreateFlag(projectID uint, key, description string, flagType models.FlagType, value interface{}) (*models.FeatureFlag, error) {
	if !flagKeyRe.MatchString(key) {
		return nil, ErrFlagKeyInvalid
	}
	if err := models.ValidateValue(flagType, value); err != nil {
		return nil, err
	}
	count, err := dal.FeatureFlag.CountByProject(projectID)
	if err != nil {
		return nil, err
	}
	if count >= MAX_FLAGS_PER_PROJECT {
		return nil, ErrFlagLimitReached
	}
	if dal.FeatureFlag.KeyExists(projectID, key) {
		return nil, ErrFlagKeyTaken
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	flag := &models.FeatureFlag{
		ProjectID:   projectID,
		Key:         key,
		Description: description,
		FlagType:    string(flagType),
		RawValue:    string(raw),
		Enabled:     true,
	}
	if err := dal.FeatureFlag.Create(flag); err != nil {
		return nil, err
	}
	return flag, nil
}

func GetFlag(projectID uint, flagID uint) (*models.FeatureFlag, error) {
	flag, err := dal.FeatureFlag.GetByID(flagID, projectID)
	if err != nil {
		return nil, ErrFlagNotFound
	}
	return flag, nil
}

func UpdateFlag(flag *models.FeatureFlag, description *string, value interface{}, enabled *bool) (*models.FeatureFlag, error) {
	if description != nil {
		flag.Description = *description
	}
	if value != nil {
		if err := models.ValidateValue(models.FlagType(flag.FlagType), value); err != nil {
			return nil, err
		}
		raw, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		flag.RawValue = string(raw)
	}
	if enabled != nil {
		flag.Enabled = *enabled
	}
	if err := dal.FeatureFlag.Save(flag); err != nil {
		return nil, err
	}
	return flag, nil
}

func DeleteFlag(flag *models.FeatureFlag) error {
	return dal.FeatureFlag.Delete(flag)
}
