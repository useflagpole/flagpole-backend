package models

import "github.com/google/uuid"

type Organization struct {
	Base
	Name    string    `gorm:"uniqueIndex;not null"`
	OwnerID uuid.UUID `gorm:"type:uuid;not null"`
	Owner   User      `gorm:"foreignKey:OwnerID"`
	Users   []User    `gorm:"many2many:auth.user_organizations;joinForeignKey:OrganizationID;joinReferences:UserID"`
}

func (Organization) TableName() string {
	return "auth.organizations"
}
