package models

import "github.com/google/uuid"

type Organization struct {
	Base
	Name    string    `gorm:"uniqueIndex;not null"                                                                  json:"name"`
	OwnerID uuid.UUID `gorm:"type:uuid;not null"                                                                    json:"ownerId"`
	Plan    string    `gorm:"not null;default:'free'"                                                               json:"plan"`
	Owner   User      `gorm:"foreignKey:OwnerID"                                                                    json:"-"`
	Users   []User    `gorm:"many2many:auth.user_organizations;joinForeignKey:OrganizationID;joinReferences:UserID" json:"-"`
}

func (Organization) TableName() string {
	return "auth.organizations"
}
