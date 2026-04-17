package database

import (
	"log"

	"flagpole/src/models"
	"flagpole/src/pkg/crypto"
)

func seedDatabase() {
	seedRoles()
	seedAdminOrg()
	seedAdminUser()
}

func seedRoles() {
	requiredRoles := []models.Role{
		{Name: "admin"},
		{Name: "editor"},
		{Name: "viewer"},
	}

	for _, role := range requiredRoles {
		var existing models.Role
		if err := DB.Where("name = ?", role.Name).First(&existing).Error; err != nil {
			if err := DB.Create(&role).Error; err != nil {
				log.Fatalf("failed to seed role '%s': %v", role.Name, err)
			}
			log.Printf("Seeded role: %s", role.Name)
		}
	}
}

func seedAdminOrg() {
	var existing models.Organization
	if err := DB.Where("name = ?", "Flagpole").First(&existing).Error; err != nil {
		org := models.Organization{Name: "Flagpole"}
		if err := DB.Create(&org).Error; err != nil {
			log.Fatalf("failed to seed admin organization: %v", err)
		}
		log.Println("Seeded organization: Flagpole")
	}
}

func seedAdminUser() {
	var adminRole models.Role
	if err := DB.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		log.Fatalf("failed to find admin role: %v", err)
	}

	var adminOrg models.Organization
	if err := DB.Where("name = ?", "Flagpole").First(&adminOrg).Error; err != nil {
		log.Fatalf("failed to find admin organization: %v", err)
	}

	var count int64
	DB.Model(&models.User{}).Where("role_id = ? AND org_id = ?", adminRole.ID, adminOrg.ID).Count(&count)
	if count > 0 {
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
		FirstName: "Admin",
		LastName:  "Admin",
		PwdHash:   hash,
		PwdSalt:   salt,
		RoleID:    adminRole.ID,
		OrgID:     adminOrg.ID,
	}

	if err := DB.Create(&admin).Error; err != nil {
		log.Fatalf("failed to create admin user: %v", err)
	}

	log.Println("-----------------------------")
	log.Println("Admin account created")
	log.Printf("Email:    %s", admin.Email)
	log.Printf("Password: %s", password)
	log.Println("-----------------------------")
}
