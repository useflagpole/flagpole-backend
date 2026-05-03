package models

type OrgRolePermission struct {
	OrgRoleID      uint   `gorm:"primaryKey;not null" json:"-"`
	PermissionCode string `gorm:"primaryKey;not null" json:"-"`
}

func (OrgRolePermission) TableName() string { return "org.org_role_permissions" }
