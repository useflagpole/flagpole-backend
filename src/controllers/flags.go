package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"flagpole/src/dal"
	"flagpole/src/database"
	"flagpole/src/models"

	"gorm.io/gorm"
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
	EnabledChanged        *bool
	RolloutToggled        *bool
	RolloutPctChanged     bool
	RolloutPct            int
	ValuesChanged         bool
	OverridesAdded        []OverrideChange
	OverridesRemoved      []OverrideChange
	OverridesValueChanged []OverrideValue
}

type OverrideChange struct {
	SegmentID   uint
	SegmentName string
}

type OverrideValue struct {
	SegmentID   uint
	SegmentName string
	OldValue    string
	NewValue    string
}

type OverridePayload struct {
	SegmentID uint        `json:"segmentId"`
	Value     interface{} `json:"value"`
	Enabled   bool        `json:"enabled"`
	Priority  int         `json:"priority"`
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
	for i, o := range overrides {
		val, _ := o.ParsedValue()
		seg, err := dal.Segment.GetByID(o.SegmentID, projectID)
		if err != nil {
			continue
		}
		overrideDTOs = append(overrideDTOs, SegmentOverrideDTO{
			ID:        o.ID,
			SegmentID: o.SegmentID,
			Name:      seg.Name,
			Value:     val,
			Enabled:   o.Enabled,
			Priority:  i + 1,
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
	Value     interface{} `json:"value"`
	Enabled   bool        `json:"enabled"`
	Priority  int         `json:"priority"`
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

func segmentName(segID uint, projectID uint) string {
	if seg, err := dal.Segment.GetByID(segID, projectID); err == nil {
		return seg.Name
	}
	return ""
}

func applyConfigChanges(
	config *models.FlagEnvironmentConfig,
	enabled *bool,
	rolloutEnabled *bool,
	rolloutPercentage *int,
	defaultValue interface{},
	servedValue interface{},
	changes *FlagConfigChanges,
) error {
	if enabled != nil && config.Enabled != *enabled {
		changes.EnabledChanged = enabled
		config.Enabled = *enabled
	}
	if rolloutEnabled != nil && config.RolloutEnabled != *rolloutEnabled {
		changes.RolloutToggled = rolloutEnabled
		config.RolloutEnabled = *rolloutEnabled
	}
	if rolloutPercentage != nil && config.RolloutPercentage != *rolloutPercentage {
		changes.RolloutPctChanged = true
		changes.RolloutPct = *rolloutPercentage
		config.RolloutPercentage = *rolloutPercentage
	}

	oldDefault := config.DefaultValue
	oldServed := config.ServedValue

	if defaultValue != nil {
		raw, err := json.Marshal(defaultValue)
		if err != nil {
			return err
		}
		config.DefaultValue = string(raw)
	}
	if servedValue != nil {
		raw, err := json.Marshal(servedValue)
		if err != nil {
			return err
		}
		config.ServedValue = string(raw)
	}
	if config.DefaultValue != oldDefault || config.ServedValue != oldServed {
		changes.ValuesChanged = true
	}
	return nil
}

// UpdateFlagConfig reconciles a flag's environment configuration and its segment overrides.
// All changes are wrapped in a transaction — if any step fails, everything rolls back.
func UpdateFlagConfig(
	projectID uint,
	flagID uint,
	env string,
	enabled *bool,
	rolloutEnabled *bool,
	rolloutPercentage *int,
	defaultValue interface{},
	servedValue interface{},
	overrides []OverridePayload,
) (*FlagConfigChanges, error) {

	var changes *FlagConfigChanges

	// Begin transaction — all DB operations below share the same tx
	err := database.DB.Transaction(func(tx *gorm.DB) error {

		// 1. Load existing config for this flag + env
		config, err := dal.FlagEnvConfig.GetByFlagAndEnv(flagID, env, tx)
		if err != nil {
			return ErrConfigNotFound
		}

		// 2. Apply base config changes (enabled, rollout, values)
		changes = &FlagConfigChanges{}
		if err := applyConfigChanges(config, enabled, rolloutEnabled, rolloutPercentage, defaultValue, servedValue, changes); err != nil {
			return err
		}

		// 3. Persist config changes
		if err := dal.FlagEnvConfig.Save(config, tx); err != nil {
			return err
		}

		// 4. No overrides provided — nothing more to do
		if overrides == nil {
			return nil
		}

		// 5. Load existing overrides for this flag + env
		existing, err := dal.FlagEnvOverride.ListByFlagAndEnv(flagID, env, tx)
		if err != nil {
			return err
		}

		// Build a map for O(1) lookup during diff
		existingMap := make(map[uint]*models.FlagEnvironmentOverride)
		for i := range existing {
			existingMap[existing[i].SegmentID] = &existing[i]
		}

		// 6. Diff incoming overrides against existing
		//    - New segments → Added
		//    - Existing segments with changed value → ValueChanged
		//    - Remaining existing segments → Removed (not in incoming list)
		for i, o := range overrides {
			raw, err := json.Marshal(o.Value)
			if err != nil {
				return err
			}
			priority := o.Priority
			if priority == 0 {
				priority = i + 1
			}

			if _, ok := existingMap[o.SegmentID]; !ok {
				// Segment not in existing overrides → new override
				if err := dal.FlagEnvOverride.SetOverride(flagID, env, o.SegmentID, string(raw), o.Enabled, priority, tx); err != nil {
					return err
				}
				changes.OverridesAdded = append(changes.OverridesAdded, OverrideChange{
					SegmentID:   o.SegmentID,
					SegmentName: segmentName(o.SegmentID, projectID),
				})
			} else {
				// Segment already has an override
				oldVal := existingMap[o.SegmentID].Value
				if string(raw) != oldVal {
					// Value changed → track it
					changes.OverridesValueChanged = append(changes.OverridesValueChanged, OverrideValue{
						SegmentID:   o.SegmentID,
						SegmentName: segmentName(o.SegmentID, projectID),
						OldValue:    oldVal,
						NewValue:    string(raw),
					})
				}
				// Always persist (updates value/enabled/priority even if value didn't change)
				if err := dal.FlagEnvOverride.SetOverride(flagID, env, o.SegmentID, string(raw), o.Enabled, priority, tx); err != nil {
					return err
				}
			}
			// Remove from map — what's left at the end are the stale overrides to delete
			delete(existingMap, o.SegmentID)
		}

		// 7. Remaining segments in existingMap were not in the incoming list → removed
		for segID := range existingMap {
			if err := dal.FlagEnvOverride.RemoveOverride(flagID, env, segID, tx); err != nil {
				return err
			}
			changes.OverridesRemoved = append(changes.OverridesRemoved, OverrideChange{
				SegmentID:   segID,
				SegmentName: segmentName(segID, projectID),
			})
		}

		return nil
	})

	return changes, err
}
