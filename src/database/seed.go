package database

import (
	"log"

	"flagpole/src/models"
	"flagpole/src/pkg/crypto"
	"flagpole/src/pkg/permissions"

	"gorm.io/gorm"
)

func seedDatabase() {
	seedAdmin()
}

// seedOrgDefaultRolesInDB creates default admin/editor/viewer roles for an org and assigns
// permissions. Mirrors dal.OrgRole.SeedForOrg — duplicated here to avoid circular imports.
// Pass a *gorm.DB as an optional second argument to run within a transaction.
func seedOrgDefaultRolesInDB(orgID uint, tx ...*gorm.DB) error {
	db := DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
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

func seedAdmin() {
	var existing models.User
	if err := DB.Where("email = ?", "admin@flagpole.dev").First(&existing).Error; err == nil {
		return
	}

	log.Println("Admin account not found, generating...")

	password, err := crypto.GenerateRandomPassword(8)
	if err != nil {
		log.Fatalf("failed to generate admin password: %v", err)
	}

	salt, err := crypto.GenerateSalt()
	if err != nil {
		log.Fatalf("failed to generate admin salt: %v", err)
	}

	hash, err := crypto.HashPassword(password, salt)
	if err != nil {
		log.Fatalf("failed to hash admin password: %v", err)
	}

	admin := models.User{
		Email:     "admin@flagpole.dev",
		Username:  "admin",
		FirstName: "Admin",
		LastName:  "Admin",
		PwdHash:   hash,
		PwdSalt:   salt,
	}

	if err := DB.Create(&admin).Error; err != nil {
		log.Fatalf("failed to create admin user: %v", err)
	}

	var org models.Organization
	err = DB.Transaction(func(tx *gorm.DB) error {
		org = models.Organization{
			Name:    "flagpole",
			OwnerID: admin.ID,
		}
		if err := tx.Create(&org).Error; err != nil {
			return err
		}
		if err := seedOrgDefaultRolesInDB(org.ID, tx); err != nil {
			return err
		}
		var adminRole models.OrgRole
		if err := tx.Where("organization_id = ? AND name = ?", org.ID, "admin").First(&adminRole).Error; err != nil {
			return err
		}
		return tx.Create(&models.UserOrganization{
			OrganizationID: org.ID,
			UserID:         admin.ID,
			OrgRoleID:      adminRole.ID,
		}).Error
	})
	if err != nil {
		log.Fatalf("failed to seed admin organization: %v", err)
	}

	log.Println("-----------------------------")
	log.Println("Admin account created")
	log.Printf("Email:    %s", admin.Email)
	log.Printf("Password: %s", password)
	log.Println("-----------------------------")
}
