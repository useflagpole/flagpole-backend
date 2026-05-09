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
		&models.Environment{},
		&models.FeatureFlag{},
		&models.FlagEnvironmentConfig{},
		&models.FlagEnvironmentOverride{},
		&models.Segment{},
		&models.SegmentRule{},
		&models.AuditLog{},
	); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	migrateFeatureFlagValues()
	migrateSegmentOverrides()

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

func migrateFeatureFlagValues() {
	// Migrate existing flag configs to flag_environment_configs table (assign to 'production' env)
	DB.Exec(`
		INSERT INTO project.flag_environment_configs 
		(flag_id, environment_name, enabled, rollout_enabled, rollout_percentage, default_value, served_value, created_at, updated_at)
		SELECT 
			ff.id, 'production', 
			COALESCE(ff.enabled, false), 
			COALESCE(ff.rollout_enabled, false), 
			COALESCE(ff.rollout_percentage, 0),
			COALESCE(ff.default_value, 'false'), 
			COALESCE(ff.served_value, 'false'),
			NOW(), NOW()
		FROM project.feature_flags ff
		WHERE ff.enabled IS NOT NULL
		ON CONFLICT DO NOTHING
	`)

	// Migrate existing segment overrides to flag_environment_overrides table (assign to 'production' env)
	DB.Exec(`
		INSERT INTO project.flag_environment_overrides
		(flag_id, environment_name, segment_id, value, enabled, created_at, updated_at)
		SELECT 
			fso.flag_id, 'production', fso.segment_id, fso.value, fso.enabled,
			NOW(), NOW()
		FROM project.flag_segment_overrides fso
		ON CONFLICT DO NOTHING
	`)

	// Drop old columns from feature_flags table after migration
	DB.Exec("ALTER TABLE project.feature_flags DROP COLUMN IF EXISTS raw_value")
	DB.Exec("ALTER TABLE project.feature_flags DROP COLUMN IF EXISTS default_value")
	DB.Exec("ALTER TABLE project.feature_flags DROP COLUMN IF EXISTS served_value")
	DB.Exec("ALTER TABLE project.feature_flags DROP COLUMN IF EXISTS rollout_enabled")
	DB.Exec("ALTER TABLE project.feature_flags DROP COLUMN IF EXISTS rollout_percentage")
	DB.Exec("ALTER TABLE project.feature_flags DROP COLUMN IF EXISTS enabled")
}

func migrateSegmentOverrides() {
	// Add match_type to segments if not exists
	DB.Exec("ALTER TABLE project.segments ADD COLUMN IF NOT EXISTS match_type VARCHAR(3) DEFAULT 'AND'")

	// Add priority to flag_environment_overrides if not exists
	DB.Exec("ALTER TABLE project.flag_environment_overrides ADD COLUMN IF NOT EXISTS priority INT DEFAULT 0")

	// Drop old flag_segment_overrides table
	DB.Exec("DROP TABLE IF EXISTS project.flag_segment_overrides")
}
