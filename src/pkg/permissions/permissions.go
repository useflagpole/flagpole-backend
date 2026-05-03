package permissions

const (
	// Feature Flags
	FlagCreate       = "flag.create"
	FlagUpdate       = "flag.update"
	FlagToggle       = "flag.toggle"
	FlagDelete       = "flag.delete"
	FlagRename       = "flag.rename"
	FlagRollout      = "flag.rollout"
	FlagRules        = "flag.rules"
	FlagArchive      = "flag.archive"

	// Segments
	SegmentCreate    = "segment.create"
	SegmentEdit      = "segment.edit"
	SegmentDelete    = "segment.delete"

	// Projects
	ProjectCreate    = "project.create"
	ProjectRename    = "project.rename"
	ProjectArchive   = "project.archive"
	ProjectUnarchive = "project.unarchive"

	// Environments
	EnvCreate        = "environment.create"
	EnvRename        = "environment.rename"
	EnvDelete        = "environment.delete"

	// Members
	MemberInvite     = "member.invite"
	MemberRemove     = "member.remove"
	MemberRole       = "member.role"

	// Organization
	OrgRename        = "org.rename"
	OrgRoles         = "org.roles"

	// SDK & Keys
	SDKView          = "sdk.view"
	SDKCreate        = "sdk.create"
	SDKRevoke        = "sdk.revoke"
)

type Def struct {
	Code        string
	Description string
}

var All = []Def{
	// Feature Flags
	{FlagCreate, "Create feature flag"},
	{FlagUpdate, "Edit flag value and description"},
	{FlagToggle, "Enable or disable flag"},
	{FlagDelete, "Delete feature flag"},
	{FlagRename, "Rename flag"},
	{FlagRollout, "Edit rollout %"},
	{FlagRules, "Edit targeting rules"},
	{FlagArchive, "Archive flag"},

	// Segments
	{SegmentCreate, "Create segment"},
	{SegmentEdit, "Edit segment rules"},
	{SegmentDelete, "Delete segment"},

	// Projects
	{ProjectCreate, "Create project"},
	{ProjectRename, "Rename project"},
	{ProjectArchive, "Archive project"},
	{ProjectUnarchive, "Unarchive project"},

	// Environments
	{EnvCreate, "Create environment"},
	{EnvRename, "Rename environment"},
	{EnvDelete, "Delete environment"},

	// Members
	{MemberInvite, "Invite member to organization"},
	{MemberRemove, "Remove member from organization"},
	{MemberRole, "Change member role"},

	// Organization
	{OrgRename, "Rename organization"},
	{OrgRoles, "Manage roles and permissions"},

	// SDK & Keys
	{SDKView, "View SDK keys"},
	{SDKCreate, "Create SDK key"},
	{SDKRevoke, "Revoke SDK key"},
}

// DefaultEditorPerms is the set of permission codes granted to the editor role by default.
var DefaultEditorPerms = map[string]bool{
	// Feature Flags
	FlagCreate: true, FlagRename: true, FlagToggle: true,
	FlagRollout: true, FlagRules: true, FlagArchive: true,

	// Segments
	SegmentCreate: true, SegmentEdit: true,

	// SDK & Keys
	SDKView: true,

	// Projects
	ProjectCreate: true, ProjectRename: true, ProjectArchive: true, ProjectUnarchive: true,

	// Environments
	EnvCreate: true, EnvRename: true, EnvDelete: true,
}

// DefaultViewerPerms is the set of permission codes granted to the viewer role by default.
var DefaultViewerPerms = map[string]bool{
	FlagToggle: true,
	SDKView:    true,
}
