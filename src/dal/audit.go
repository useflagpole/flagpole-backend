package dal

import (
	"time"

	"flagpole/src/database"
	"flagpole/src/models"

	"github.com/google/uuid"
)

type auditDAL struct{}

var Audit = auditDAL{}

type AuditEntry struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	OrgID     uint      `json:"orgId"`
	ProjectID *uint     `json:"projectId"`
	Actor     string    `json:"actor"`
	Action    string    `json:"action"`
	Target    string    `json:"target"`
	Detail    string    `json:"detail"`
	Env       string    `json:"env"`
}

func (auditDAL) Log(entry models.AuditLog) error {
	return database.DB.Create(&entry).Error
}

func (auditDAL) ListByProject(projectID uint) ([]AuditEntry, error) {
	var entries []AuditEntry
	err := database.DB.Raw(`
		SELECT al.id, al.created_at, al.org_id, al.project_id,
		       COALESCE(u.username, al.actor_email) AS actor,
		       al.action, al.target, al.detail, al.env
		FROM audit.audit_logs al
		LEFT JOIN auth.users u ON u.id = al.actor_id
		WHERE al.project_id = ?
		ORDER BY al.created_at DESC
	`, projectID).Scan(&entries).Error
	return entries, err
}

func (auditDAL) ListByTarget(projectID uint, target string, env string) ([]AuditEntry, error) {
	var entries []AuditEntry
	query := `
		SELECT al.id, al.created_at, al.org_id, al.project_id,
		       COALESCE(u.username, al.actor_email) AS actor,
		       al.action, al.target, al.detail, al.env
		FROM audit.audit_logs al
		LEFT JOIN auth.users u ON u.id = al.actor_id
		WHERE al.project_id = ? AND al.target = ?`
	args := []interface{}{projectID, target}
	if env != "" {
		query += ` AND (al.env = ? OR al.env = '' OR al.env IS NULL)`
		args = append(args, env)
	}
	query += ` ORDER BY al.created_at DESC`
	err := database.DB.Raw(query, args...).Scan(&entries).Error
	return entries, err
}

func (auditDAL) ListByOrg(orgID uint) ([]AuditEntry, error) {
	var entries []AuditEntry
	err := database.DB.Raw(`
		SELECT al.id, al.created_at, al.org_id, al.project_id,
		       COALESCE(u.username, al.actor_email) AS actor,
		       al.action, al.target, al.detail, al.env
		FROM audit.audit_logs al
		LEFT JOIN auth.users u ON u.id = al.actor_id
		WHERE al.org_id = ?
		ORDER BY al.created_at DESC
	`, orgID).Scan(&entries).Error
	return entries, err
}

func (auditDAL) LogWithActor(orgID uint, projectID *uint, actorID uuid.UUID, action, target, detail, env string) error {
	var email string
	database.DB.Raw(`SELECT email FROM auth.users WHERE id = ?`, actorID).Scan(&email)
	entry := models.AuditLog{
		OrgID:      orgID,
		ProjectID:  projectID,
		ActorID:    actorID,
		ActorEmail: email,
		Action:     action,
		Target:     target,
		Detail:     detail,
		Env:        env,
	}
	return database.DB.Create(&entry).Error
}
