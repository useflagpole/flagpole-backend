package models

import "github.com/google/uuid"

type UserOrganization struct {
	UserID         uuid.UUID `gorm:"primaryKey;type:uuid"`
	OrganizationID uint      `gorm:"primaryKey"`
}

func (UserOrganization) TableName() string {
	return "auth.user_organizations"
}
