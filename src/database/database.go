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
	if err := DB.AutoMigrate(&models.Role{}, &models.Organization{}, &models.User{}, &models.UserOrganization{}, &models.FeatureFlag{}); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	dropLegacyColumns()
	ensureOwnerMembershipTrigger()
	log.Println("Migrations applied")
}

func ensureOwnerMembershipTrigger() {
	statements := []string{
		`CREATE OR REPLACE FUNCTION auth.check_owner_is_member()
		RETURNS TRIGGER LANGUAGE plpgsql AS $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM auth.user_organizations
				WHERE user_id = NEW.owner_id
				AND organization_id = NEW.id
			) THEN
				RAISE EXCEPTION 'owner must be a member of the organization';
			END IF;
			RETURN NEW;
		END;
		$$`,

		`DROP TRIGGER IF EXISTS trg_owner_is_member ON auth.organizations`,

		`CREATE CONSTRAINT TRIGGER trg_owner_is_member
		AFTER INSERT OR UPDATE OF owner_id ON auth.organizations
		DEFERRABLE INITIALLY DEFERRED
		FOR EACH ROW EXECUTE FUNCTION auth.check_owner_is_member()`,
	}

	for _, stmt := range statements {
		if err := DB.Exec(stmt).Error; err != nil {
			log.Fatalf("failed to create owner membership trigger: %v", err)
		}
	}
}

func dropLegacyColumns() {
	migrator := DB.Migrator()
	if migrator.HasColumn(&models.User{}, "org_id") {
		if err := migrator.DropColumn(&models.User{}, "org_id"); err != nil {
			log.Fatalf("failed to drop legacy org_id column: %v", err)
		}
		log.Println("Dropped legacy column: auth.users.org_id")
	}
}
