package models

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

const ENV_NAME_MIN_LEN = 2
const ENV_NAME_MAX_LEN = 32

type Environment struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	ProjectID uint   `gorm:"not null;uniqueIndex:idx_env_project_name" json:"projectId"`
	Name      string `gorm:"not null;uniqueIndex:idx_env_project_name" json:"name"`
}

func (Environment) TableName() string {
	return "project.environments"
}

func ValidateEnvironmentName(name string) error {
	name = strings.TrimSpace(name)
	if len(name) < ENV_NAME_MIN_LEN {
		return errors.New("environment name must be at least 2 characters")
	}
	if len(name) > ENV_NAME_MAX_LEN {
		return errors.New("environment name must be at most 32 characters")
	}
	if !regexp.MustCompile(`^[a-z0-9-]+$`).MatchString(name) {
		return errors.New("environment name must contain only lowercase letters, numbers, and hyphens")
	}
	if name[0] == '-' || name[len(name)-1] == '-' {
		return errors.New("environment name cannot start or end with hyphen")
	}
	return nil
}
