package database

import (
	"log"

	"flagpole/src/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(dsn string) error {
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return err
	}
	migrate()
	seedDatabase()
	return nil
}

func migrate() {
	if err := DB.Exec("CREATE SCHEMA IF NOT EXISTS auth").Error; err != nil {
		log.Fatalf("failed to create auth schema: %v", err)
	}
	if err := DB.AutoMigrate(&models.Role{}, &models.Organization{}, &models.User{}, &models.FeatureFlag{}); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	log.Println("Migrations applied")
}
