package models

type OrgRole struct {
	ID             uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	OrganizationID uint   `gorm:"not null;index"           json:"organizationId"`
	Name           string `gorm:"not null"                 json:"name"`
	IsDefault      bool   `gorm:"not null;default:false"   json:"isDefault"`
	IsProtected    bool   `gorm:"not null;default:false"   json:"isProtected"`
}

func (OrgRole) TableName() string { return "org.org_roles" }
