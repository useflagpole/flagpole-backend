package controllers

import (
	"encoding/json"
	"errors"
	"regexp"

	"flagpole/src/dal"
	"flagpole/src/models"
)

const MAX_FLAGS_PER_PROJECT = 25
const FLAG_NAME_MAX_LEN     = 50

var (
	ErrFlagKeyInvalid   = errors.New("key must be 2–64 chars, lowercase letters, numbers, hyphens or underscores")
	ErrFlagKeyTaken     = errors.New("a flag with that key already exists in this project")
	ErrFlagNotFound     = errors.New("flag not found")
	ErrFlagNameRequired = errors.New("name is required")
	ErrFlagNameInvalid  = errors.New("name must be ≤50 chars, alphanumeric and hyphens only, cannot start with a hyphen")
	ErrFlagLimitReached = errors.New("project has reached the maximum of 25 flags")
)

var flagKeyRe  = regexp.MustCompile(`^[a-z0-9_-]{2,64}$`)
var flagNameRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{0,48}[a-zA-Z0-9]$|^[a-zA-Z0-9]$`)

func sanitizeFlagName(name string) string {
	// replace spaces with hyphens
	result := regexp.MustCompile(`\s+`).ReplaceAllString(name, "-")
	// strip disallowed chars (keep alphanumeric and hyphens)
	result = regexp.MustCompile(`[^a-zA-Z0-9-]`).ReplaceAllString(result, "")
	return result
}

func validateFlagName(name string) error {
	if name == "" {
		return ErrFlagNameRequired
	}
	if len(name) > FLAG_NAME_MAX_LEN {
		return ErrFlagNameInvalid
	}
	if !flagNameRe.MatchString(name) {
		return ErrFlagNameInvalid
	}
	return nil
}

func ListFlags(projectID uint) ([]models.FeatureFlag, error) {
	return dal.FeatureFlag.ListByProject(projectID)
}

func CreateFlag(projectID uint, key, name string, flagType models.FlagType, value interface{}) (*models.FeatureFlag, error) {
	if !flagKeyRe.MatchString(key) {
		return nil, ErrFlagKeyInvalid
	}
	name = sanitizeFlagName(name)
	if err := validateFlagName(name); err != nil {
		return nil, err
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
		ProjectID: projectID,
		Key:       key,
		Name:      name,
		FlagType:  string(flagType),
		RawValue:  string(raw),
		Enabled:   true,
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

func UpdateFlag(flag *models.FeatureFlag, name *string, value interface{}, enabled *bool) (*models.FeatureFlag, error) {
	if name != nil {
		sanitized := sanitizeFlagName(*name)
		if err := validateFlagName(sanitized); err != nil {
			return nil, err
		}
		flag.Name = sanitized
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
