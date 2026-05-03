package dal

import (
	"errors"

	"flagpole/src/database"
	"flagpole/src/models"
	"flagpole/src/pkg/permissions"

	"gorm.io/gorm"
)

type orgRoleDAL struct{}

var OrgRole = orgRoleDAL{}

func (orgRoleDAL) db(tx ...*gorm.DB) *gorm.DB {
	if len(tx) > 0 && tx[0] != nil {
		return tx[0]
	}
	return database.DB
}

func (orgRoleDAL) GetByID(id uint) (*models.OrgRole, error) {
	var role models.OrgRole
	if err := database.DB.First(&role, id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (orgRoleDAL) GetAdminRole(orgID uint, tx ...*gorm.DB) (*models.OrgRole, error) {
	var role models.OrgRole
	err := OrgRole.db(tx...).Where("organization_id = ? AND name = ?", orgID, "admin").First(&role).Error
	return &role, err
}

func (orgRoleDAL) ListByOrg(orgID uint) ([]models.OrgRole, error) {
	var roles []models.OrgRole
	err := database.DB.Where("organization_id = ?", orgID).Find(&roles).Error
	return roles, err
}

func (orgRoleDAL) Create(role *models.OrgRole) error {
	return database.DB.Create(role).Error
}

func (orgRoleDAL) Delete(role *models.OrgRole) error {
	if role.IsProtected {
		return errors.New("cannot delete protected role")
	}
	return database.DB.Delete(role).Error
}

func (orgRoleDAL) GetPermissions(orgRoleID uint) ([]string, error) {
	var codes []string
	err := database.DB.
		Model(&models.OrgRolePermission{}).
		Where("org_role_id = ?", orgRoleID).
		Pluck("permission_code", &codes).Error
	return codes, err
}

func (orgRoleDAL) SetPermission(orgRoleID uint, code string, enabled bool) error {
	if enabled {
		return database.DB.Exec(`
			INSERT INTO org.org_role_permissions (org_role_id, permission_code)
			VALUES (?, ?)
			ON CONFLICT DO NOTHING
		`, orgRoleID, code).Error
	}
	return database.DB.
		Where("org_role_id = ? AND permission_code = ?", orgRoleID, code).
		Delete(&models.OrgRolePermission{}).Error
}

// SeedForOrg creates the default admin/editor/viewer roles for an org and assigns
// their permissions. Idempotent — safe to call on existing orgs.
// Pass a *gorm.DB as an optional second argument to run within a transaction.
func (orgRoleDAL) SeedForOrg(orgID uint, tx ...*gorm.DB) error {
	db := OrgRole.db(tx...)

	type roleSpec struct {
		name        string
		isProtected bool
		perms       map[string]bool
	}

	allPerms := make(map[string]bool, len(permissions.All))
	for _, p := range permissions.All {
		allPerms[p.Code] = true
	}

	specs := []roleSpec{
		{name: "admin", isProtected: true, perms: allPerms},
		{name: "editor", isProtected: false, perms: permissions.DefaultEditorPerms},
		{name: "viewer", isProtected: false, perms: map[string]bool{}},
	}

	for _, spec := range specs {
		var role models.OrgRole
		result := db.Where("organization_id = ? AND name = ?", orgID, spec.name).First(&role)
		if result.Error != nil {
			role = models.OrgRole{
				OrganizationID: orgID,
				Name:           spec.name,
				IsDefault:      true,
				IsProtected:    spec.isProtected,
			}
			if err := db.Create(&role).Error; err != nil {
				return err
			}
		}

		for code := range spec.perms {
			db.Exec(`
				INSERT INTO org.org_role_permissions (org_role_id, permission_code)
				VALUES (?, ?) ON CONFLICT DO NOTHING
			`, role.ID, code)
		}
	}
	return nil
}
