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
	backfillOrganizationOwner()
	backfillUserOrgRole()
	backfillUsername()
	backfillFlagDescription()
	backfillAuditLogActorIDType()
	if err := DB.AutoMigrate(&models.Role{}, &models.Organization{}, &models.User{}, &models.UserOrganization{}, &models.Project{}, &models.FeatureFlag{}, &models.AuditLog{}); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	dropLegacyColumns()
	ensureOwnerMembershipTrigger()
	log.Println("Migrations applied")
}

func backfillOrganizationOwner() {
	if DB.Migrator().HasColumn(&models.Organization{}, "owner_id") {
		return
	}

	steps := []string{
		`ALTER TABLE auth.organizations ADD COLUMN owner_id uuid`,
		`UPDATE auth.organizations
		 SET owner_id = (SELECT id FROM auth.users WHERE email = 'admin@flagpole.dev')
		 WHERE owner_id IS NULL`,
		`ALTER TABLE auth.organizations ALTER COLUMN owner_id SET NOT NULL`,
	}

	for _, stmt := range steps {
		if err := DB.Exec(stmt).Error; err != nil {
			log.Fatalf("backfillOrganizationOwner: %v", err)
		}
	}

	log.Println("Backfilled organizations.owner_id from admin user")
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

func backfillUserOrgRole() {
	if DB.Migrator().HasColumn(&models.UserOrganization{}, "role_id") {
		return
	}

	steps := []string{
		`ALTER TABLE auth.user_organizations ADD COLUMN role_id bigint`,
		`UPDATE auth.user_organizations
		 SET role_id = (SELECT id FROM auth.roles WHERE name = 'viewer')
		 WHERE role_id IS NULL`,
		`ALTER TABLE auth.user_organizations ALTER COLUMN role_id SET NOT NULL`,
	}

	for _, stmt := range steps {
		if err := DB.Exec(stmt).Error; err != nil {
			log.Fatalf("backfillUserOrgRole: %v", err)
		}
	}

	log.Println("Backfilled user_organizations.role_id to viewer")
}

func backfillUsername() {
	if DB.Migrator().HasColumn(&models.User{}, "username") {
		return
	}

	steps := []string{
		`ALTER TABLE auth.users ADD COLUMN username text`,
		`UPDATE auth.users
		 SET username = lower(regexp_replace(first_name || last_name || '_' || substr(id::text, 1, 6), '[^a-z0-9_]', '', 'g'))
		 WHERE username IS NULL`,
		`ALTER TABLE auth.users ALTER COLUMN username SET NOT NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON auth.users(username)`,
	}

	for _, stmt := range steps {
		if err := DB.Exec(stmt).Error; err != nil {
			log.Fatalf("backfillUsername: %v", err)
		}
	}

	log.Println("Backfilled auth.users.username")
}

func backfillAuditLogActorIDType() {
	var colType string
	DB.Raw(`
		SELECT data_type FROM information_schema.columns
		WHERE table_schema = 'auth' AND table_name = 'audit_logs' AND column_name = 'actor_id'
	`).Scan(&colType)
	if colType == "" || colType == "uuid" {
		return
	}
	if err := DB.Exec(`ALTER TABLE auth.audit_logs ALTER COLUMN actor_id TYPE uuid USING actor_id::uuid`).Error; err != nil {
		log.Fatalf("backfillAuditLogActorIDType: %v", err)
	}
	log.Println("Converted audit_logs.actor_id to uuid type")
}

func backfillFlagDescription() {
	if DB.Migrator().HasColumn(&models.FeatureFlag{}, "description") {
		return
	}
	steps := []string{
		`ALTER TABLE auth.feature_flags RENAME COLUMN name TO description`,
	}
	for _, stmt := range steps {
		if err := DB.Exec(stmt).Error; err != nil {
			log.Fatalf("backfillFlagDescription: %v", err)
		}
	}
	log.Println("Renamed feature_flags.name to description")
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
