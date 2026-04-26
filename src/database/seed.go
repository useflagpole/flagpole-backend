package database

import (
	"log"

	"flagpole/src/models"
	"flagpole/src/pkg/crypto"

	"gorm.io/gorm"
)

func seedDatabase() {
	seedRoles()
	seedAdmin()
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

func seedAdmin() {
	var existing models.User
	if err := DB.Where("email = ?", "admin@flagpole.dev").First(&existing).Error; err == nil {
		return
	}

	log.Println("Admin account not found, generating...")

	var adminRole models.Role
	if err := DB.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		log.Fatalf("failed to find admin role: %v", err)
	}

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
	}

	if err := DB.Create(&admin).Error; err != nil {
		log.Fatalf("failed to create admin user: %v", err)
	}

	err = DB.Transaction(func(tx *gorm.DB) error {
		org := models.Organization{
			Name:    "flagpole",
			OwnerID: admin.ID,
		}
		if err := tx.Create(&org).Error; err != nil {
			return err
		}
		return tx.Create(&models.UserOrganization{
			OrganizationID: org.ID,
			UserID:         admin.ID,
			RoleID:         adminRole.ID,
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
