package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	SDKKeyTypeServer   = "server"
	SDKKeyTypeClient   = "client"
	SDKKeyPrefixServer = "fp_srv_"
	SDKKeyPrefixClient = "fp_cli_"
	SDKKeyNameMaxLen   = 64
)

type SDKKey struct {
	ID            uint           `gorm:"primaryKey;autoIncrement"                   json:"id"`
	CreatedAt     time.Time      `                                                  json:"createdAt"`
	UpdatedAt     time.Time      `                                                  json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index"                                      json:"-"`
	ProjectID     uint           `gorm:"not null;index"                             json:"projectId"`
	EnvironmentID uint           `gorm:"not null;index"                             json:"environmentId"`
	Environment   Environment    `gorm:"foreignKey:EnvironmentID"                   json:"-"`
	KeyType       string         `gorm:"not null"                                   json:"type"`
	Name          string         `gorm:"not null"                                   json:"name"`
	Key           string         `gorm:"not null;uniqueIndex"                       json:"-"`
	RevokedAt     *time.Time     `                                                  json:"revokedAt"`
	LastUsedAt    *time.Time     `                                                  json:"lastUsedAt"`
}

func (SDKKey) TableName() string { return "project.sdk_keys" }
