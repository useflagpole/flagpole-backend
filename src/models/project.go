package models

type Project struct {
	Base
	Name           string `gorm:"not null"                                                   json:"name"`
	OrganizationID uint   `gorm:"not null"                                                   json:"organizationId"`
	Environments   string `gorm:"not null;default:'[\"production\",\"staging\",\"dev\"]'"    json:"environments"`
}

func (Project) TableName() string {
	return "auth.projects"
}
