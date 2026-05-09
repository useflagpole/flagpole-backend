package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	ActionFlagCreate  = "flag.create"
	ActionFlagToggle  = "flag.toggle"
	ActionFlagUpdate  = "flag.update"
	ActionFlagDelete  = "flag.delete"
	ActionFlagRollout = "flag.rollout"
	ActionFlagValues  = "flag.values"

	ActionSegmentCreate       = "segment.create"
	ActionSegmentUpdate       = "segment.update"
	ActionSegmentDelete       = "segment.delete"
	ActionFlagOverrideAdd     = "flag.override_add"
	ActionFlagOverrideRemove  = "flag.override_remove"
	ActionFlagOverrideUpdate  = "flag.override_update"

	ActionProjectCreate    = "project.create"
	ActionProjectRename    = "project.rename"
	ActionProjectArchive   = "project.archive"
	ActionProjectUnarchive = "project.unarchive"

	ActionOrgCreate = "org.create"
	ActionOrgRename = "org.rename"
	ActionOrgInvite = "org.invite"
	ActionOrgJoin   = "org.join"

	ActionRoleCreate      = "role.create"
	ActionRoleDelete      = "role.delete"
	ActionRoleUpdatePerms = "role.update_perms"
	ActionMemberRole      = "member.role"

	ActionEnvCreate = "env.create"
	ActionEnvRename = "env.rename"
	ActionEnvDelete = "env.delete"

	ActionOrgDelete = "org.delete"
	ActionOrgPlan   = "org.plan"
)

type AuditLog struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	OrgID     uint  `gorm:"not null;index:idx_audit_org"     json:"orgId"`
	ProjectID *uint `gorm:"index:idx_audit_project"          json:"projectId"`

	ActorID    uuid.UUID `gorm:"type:uuid;not null" json:"-"`
	ActorEmail string    `gorm:"not null"           json:"-"`

	Action string `gorm:"not null"     json:"action"`
	Target string `gorm:"not null"     json:"target"`
	Detail string `gorm:"not null"     json:"detail"`
	Env    string `gorm:"default:null" json:"env"`
}

func (AuditLog) TableName() string { return "audit.audit_logs" }
