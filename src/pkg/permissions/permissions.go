package permissions

const (
	FlagCreate       = "flag.create"
	FlagUpdate       = "flag.update"
	FlagToggle       = "flag.toggle"
	FlagDelete       = "flag.delete"
	ProjectCreate    = "project.create"
	ProjectRename    = "project.rename"
	ProjectArchive   = "project.archive"
	ProjectUnarchive = "project.unarchive"
	EnvCreate        = "environment.create"
	EnvRename        = "environment.rename"
	EnvDelete        = "environment.delete"
	MemberInvite     = "member.invite"
	MemberRemove     = "member.remove"
	MemberRole       = "member.role"
	OrgRename        = "org.rename"
	OrgRoles         = "org.roles"
)

type Def struct {
	Code        string
	Description string
}

var All = []Def{
	{FlagCreate, "Create feature flag"},
	{FlagUpdate, "Edit flag value and description"},
	{FlagToggle, "Enable or disable flag"},
	{FlagDelete, "Delete feature flag"},
	{ProjectCreate, "Create project"},
	{ProjectRename, "Rename project"},
	{ProjectArchive, "Archive project"},
	{ProjectUnarchive, "Unarchive project"},
	{EnvCreate, "Create environment"},
	{EnvRename, "Rename environment"},
	{EnvDelete, "Delete environment"},
	{MemberInvite, "Invite member to organization"},
	{MemberRemove, "Remove member from organization"},
	{MemberRole, "Change member role"},
	{OrgRename, "Rename organization"},
	{OrgRoles, "Manage roles and permissions"},
}

// DefaultEditorPerms is the set of permission codes granted to the editor role by default.
var DefaultEditorPerms = map[string]bool{
	FlagCreate: true, FlagUpdate: true, FlagToggle: true, FlagDelete: true,
	ProjectCreate: true, ProjectRename: true, ProjectArchive: true, ProjectUnarchive: true,
	EnvCreate: true, EnvRename: true, EnvDelete: true,
}
