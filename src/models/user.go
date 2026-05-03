package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	Base
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Email         string         `gorm:"uniqueIndex;not null"`
	Username      string         `gorm:"uniqueIndex;not null"`
	FirstName     string         `gorm:"not null"`
	LastName      string         `gorm:"not null"`
	PwdHash       string         `gorm:"not null"`
	PwdSalt       string         `gorm:"not null"`
	Organizations []Organization `gorm:"many2many:org.user_organizations;joinForeignKey:UserID;joinReferences:OrganizationID"`
}

func (User) TableName() string {
	return "auth.users" // auth schema: identity only
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
