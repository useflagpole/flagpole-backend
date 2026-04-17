package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	Base
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email     string    `gorm:"uniqueIndex;not null"`
	FirstName string    `gorm:"not null"`
	LastName  string    `gorm:"not null"`
	PwdHash   string    `gorm:"not null"`
	PwdSalt   string    `gorm:"not null"`
	RoleID    uint      `gorm:"not null;constraint:OnDelete:RESTRICT"`
	OrgID     uint      `gorm:"not null;constraint:OnDelete:RESTRICT"`
}

func (User) TableName() string {
	return "auth.users"
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
