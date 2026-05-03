package models

import "github.com/google/uuid"

type UserOrganization struct {
	UserID         uuid.UUID `gorm:"primaryKey;type:uuid"`
	OrganizationID uint      `gorm:"primaryKey"`
	OrgRoleID      uint      `gorm:"not null"`
}

func (UserOrganization) TableName() string {
	return "org.user_organizations"
}
