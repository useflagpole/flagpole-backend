package models

type Organization struct {
	Base
	Name string `gorm:"uniqueIndex;not null"`
}

func (Organization) TableName() string {
	return "auth.organizations"
}
