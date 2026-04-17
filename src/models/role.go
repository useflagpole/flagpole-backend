package models

type Role struct {
	Base
	Name string `gorm:"uniqueIndex;not null"`
}

func (Role) TableName() string {
	return "auth.roles"
}
