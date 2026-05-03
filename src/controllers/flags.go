package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"flagpole/src/dal"
	"flagpole/src/models"
)

const MAX_FLAGS_PER_PROJECT = 25
const FLAG_KEY_MIN_LEN = 2
const FLAG_KEY_MAX_LEN = 64

var (
	ErrFlagKeyInvalid   = errors.New("key must be 2–64 chars, lowercase letters, numbers, hyphens or underscores")
	ErrFlagKeyTaken     = errors.New("a flag with that key already exists in this project")
	ErrFlagNotFound     = errors.New("flag not found")
	ErrFlagLimitReached = errors.New("project has reached the maximum of " + fmt.Sprint(MAX_FLAGS_PER_PROJECT) + " flags")
	ErrConfigNotFound   = errors.New("configuration not found for this environment")
	ErrConfigExists     = errors.New("configuration already exists for this environment")
)

var flagKeyRe = regexp.MustCompile(`^[a-z0-9_-]{` + fmt.Sprint(FLAG_KEY_MIN_LEN) + `,` + fmt.Sprint(FLAG_KEY_MAX_LEN) + `}$`)

func ListFlags(projectID uint) ([]models.FeatureFlag, error) {
	return dal.FeatureFlag.ListByProject(projectID)
}

func GetFlagAudit(projectID uint, flagKey string, env string) ([]dal.AuditEntry, error) {
	return dal.Audit.ListByTarget(projectID, flagKey, env)
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
	flag := &models.FeatureFlag{
		ProjectID:   projectID,
		Key:         key,
		Description: description,
		FlagType:    string(flagType),
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

func DeleteFlag(flag *models.FeatureFlag) error {
	return dal.FeatureFlag.Delete(flag)
}

func UpdateFlagMetadata(flag *models.FeatureFlag, description *string) (*models.FeatureFlag, error) {
	if description != nil {
		flag.Description = *description
		if err := dal.FeatureFlag.Save(flag); err != nil {
			return nil, err
		}
	}
	return flag, nil
}

type FlagConfigChanges struct {
	EnabledChanged     *bool
	RolloutToggled     *bool
	RolloutPctChanged  bool
	RolloutPct         int
	ValuesChanged      bool
	OverridesAdded     []uint
	OverridesRemoved   []uint
}

type OverridePayload struct {
	SegmentID uint        `json:"segmentId"`
	Value     interface{} `json:"value"`
	Enabled   bool        `json:"enabled"`
}

func GetFlagDetail(projectID uint, flagID uint, env string) (*FlagDetailDTO, error) {
	flag, err := dal.FeatureFlag.GetByID(flagID, projectID)
	if err != nil {
		return nil, ErrFlagNotFound
	}

	config, err := dal.FlagEnvConfig.GetByFlagAndEnv(flagID, env)
	if err != nil {
		return &FlagDetailDTO{
			ID:               flag.ID,
			Key:              flag.Key,
			Type:             flag.FlagType,
			Description:      flag.Description,
			Status:           "",
			Rollout:          0,
			RolloutEnabled:   false,
			DefaultValue:     nil,
			ServedValue:      nil,
			SegmentOverrides: []SegmentOverrideDTO{},
		}, nil
	}

	defaultVal, _ := config.ParsedDefaultValue()
	servedVal, _ := config.ParsedServedValue()

	overrides, err := dal.FlagEnvOverride.ListByFlagAndEnv(flagID, env)
	if err != nil {
		return nil, err
	}

	overrideDTOs := make([]SegmentOverrideDTO, 0, len(overrides))
	for _, o := range overrides {
		val, _ := o.ParsedValue()
		seg, err := dal.Segment.GetByID(o.SegmentID, projectID)
		if err != nil {
			continue
		}
		overrideDTOs = append(overrideDTOs, SegmentOverrideDTO{
			ID:        o.ID,
			SegmentID: o.SegmentID,
			Name:      seg.Name,
			UserCount: seg.UserCount,
			Value:     val,
			Enabled:   o.Enabled,
		})
	}

	status := "off"
	if config.Enabled {
		status = "on"
	}

	return &FlagDetailDTO{
		ID:               flag.ID,
		Key:              flag.Key,
		Type:             flag.FlagType,
		Description:      flag.Description,
		Status:           status,
		Rollout:          config.RolloutPercentage,
		RolloutEnabled:   config.RolloutEnabled,
		DefaultValue:     defaultVal,
		ServedValue:      servedVal,
		SegmentOverrides: overrideDTOs,
	}, nil
}

type FlagDetailDTO struct {
	ID               uint                 `json:"id"`
	Key              string               `json:"key"`
	Type             string               `json:"type"`
	Description      string               `json:"description"`
	Status           string               `json:"status"`
	Rollout          int                  `json:"rollout"`
	RolloutEnabled   bool                 `json:"rolloutEnabled"`
	DefaultValue     interface{}          `json:"defaultValue"`
	ServedValue      interface{}          `json:"servedValue"`
	SegmentOverrides []SegmentOverrideDTO `json:"segmentOverrides"`
}

type SegmentOverrideDTO struct {
	ID        uint        `json:"id"`
	SegmentID uint        `json:"segmentId"`
	Name      string      `json:"name"`
	UserCount int         `json:"userCount"`
	Value     interface{} `json:"value"`
	Enabled   bool        `json:"enabled"`
}

func CreateFlagEnvConfig(flagID uint, env string, projectID uint, flagType string) (*models.FlagEnvironmentConfig, error) {
	if dal.FlagEnvConfig.Exists(flagID, env) {
		return nil, ErrConfigExists
	}

	defaultValue := defaultValueForType(flagType)
	servedValue := defaultValueForType(flagType)

	config := &models.FlagEnvironmentConfig{
		FlagID:            flagID,
		EnvironmentName:   env,
		Enabled:           false,
		RolloutEnabled:    false,
		RolloutPercentage: 0,
		DefaultValue:      defaultValue,
		ServedValue:       servedValue,
	}

	if err := dal.FlagEnvConfig.Create(config); err != nil {
		return nil, err
	}

	return config, nil
}

func defaultValueForType(flagType string) string {
	switch flagType {
	case "bool":
		return "false"
	case "number":
		return "0"
	case "string":
		return ""
	default:
		return "false"
	}
}

func UpdateFlagConfig(flagID uint, env string,
	enabled *bool, rolloutEnabled *bool, rolloutPercentage *int,
	defaultValue interface{}, servedValue interface{},
	overrides []OverridePayload) (*FlagConfigChanges, error) {

	config, err := dal.FlagEnvConfig.GetByFlagAndEnv(flagID, env)
	if err != nil {
		return nil, ErrConfigNotFound
	}

	changes := &FlagConfigChanges{}

	oldEnabled := config.Enabled
	if enabled != nil {
		config.Enabled = *enabled
		if config.Enabled != oldEnabled {
			changes.EnabledChanged = &config.Enabled
		}
	}

	oldRolloutEnabled := config.RolloutEnabled
	oldRolloutPct := config.RolloutPercentage
	if rolloutEnabled != nil {
		config.RolloutEnabled = *rolloutEnabled
		if config.RolloutEnabled != oldRolloutEnabled {
			changes.RolloutToggled = &config.RolloutEnabled
		}
	}
	if rolloutPercentage != nil {
		config.RolloutPercentage = *rolloutPercentage
		if config.RolloutPercentage != oldRolloutPct {
			changes.RolloutPctChanged = true
			changes.RolloutPct = config.RolloutPercentage
		}
	}

	oldDefault := config.DefaultValue
	oldServed := config.ServedValue
	if defaultValue != nil {
		raw, err := json.Marshal(defaultValue)
		if err != nil {
			return nil, err
		}
		config.DefaultValue = string(raw)
	}
	if servedValue != nil {
		raw, err := json.Marshal(servedValue)
		if err != nil {
			return nil, err
		}
		config.ServedValue = string(raw)
	}
	if config.DefaultValue != oldDefault || config.ServedValue != oldServed {
		changes.ValuesChanged = true
	}

	if err := dal.FlagEnvConfig.Save(config); err != nil {
		return nil, err
	}

	if overrides != nil {
		existing, err := dal.FlagEnvOverride.ListByFlagAndEnv(flagID, env)
		if err != nil {
			return nil, err
		}
		existingMap := make(map[uint]bool)
		for _, e := range existing {
			existingMap[e.SegmentID] = true
		}

		for _, o := range overrides {
			raw, err := json.Marshal(o.Value)
			if err != nil {
				return nil, err
			}
			if err := dal.FlagEnvOverride.SetOverride(flagID, env, o.SegmentID, string(raw), o.Enabled); err != nil {
				return nil, err
			}
			if !existingMap[o.SegmentID] {
				changes.OverridesAdded = append(changes.OverridesAdded, o.SegmentID)
			}
			delete(existingMap, o.SegmentID)
		}

		for segID := range existingMap {
			if err := dal.FlagEnvOverride.RemoveOverride(flagID, env, segID); err != nil {
				return nil, err
			}
			changes.OverridesRemoved = append(changes.OverridesRemoved, segID)
		}
	}

	return changes, nil
}
