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
	for _, schema := range []string{"auth", "org", "project", "audit"} {
		if err := DB.Exec("CREATE SCHEMA IF NOT EXISTS " + schema).Error; err != nil {
			log.Fatalf("failed to create schema %s: %v", schema, err)
		}
	}

	if err := DB.AutoMigrate(
		&models.User{},
		&models.Organization{},
		&models.UserOrganization{},
		&models.OrgRole{},
		&models.OrgRolePermission{},
		&models.Project{},
		&models.FeatureFlag{},
		&models.AuditLog{},
	); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	ensureOwnerMembershipTrigger()
	log.Println("Migrations applied")
}

func ensureOwnerMembershipTrigger() {
	statements := []string{
		`CREATE OR REPLACE FUNCTION org.check_owner_is_member()
		RETURNS TRIGGER LANGUAGE plpgsql AS $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM org.user_organizations
				WHERE user_id = NEW.owner_id
				AND organization_id = NEW.id
			) THEN
				RAISE EXCEPTION 'owner must be a member of the organization';
			END IF;
			RETURN NEW;
		END;
		$$`,

		`DROP TRIGGER IF EXISTS trg_owner_is_member ON org.organizations`,

		`CREATE CONSTRAINT TRIGGER trg_owner_is_member
		AFTER INSERT OR UPDATE OF owner_id ON org.organizations
		DEFERRABLE INITIALLY DEFERRED
		FOR EACH ROW EXECUTE FUNCTION org.check_owner_is_member()`,
	}

	for _, stmt := range statements {
		if err := DB.Exec(stmt).Error; err != nil {
			log.Fatalf("failed to create owner membership trigger: %v", err)
		}
	}
}
