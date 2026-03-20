package migrate

// MigrateConfig is the portable format for cross-machine migration.
type MigrateConfig struct {
	Version    string    `json:"version"`
	Timestamp  string    `json:"timestamp"`
	SourceOS   string    `json:"source_os"`
	SourceHome string    `json:"source_home"`
	SSH        SSHExport `json:"ssh"`
	Git        GitExport `json:"git"`
}

// SSHExport holds portable SSH configuration.
type SSHExport struct {
	Hosts []SSHHostEntry `json:"hosts"`
	Keys  []SSHKeyRef    `json:"keys"`
}

// SSHHostEntry represents a single SSH host for migration.
type SSHHostEntry struct {
	Name         string `json:"name"`
	Hostname     string `json:"hostname,omitempty"`
	User         string `json:"user,omitempty"`
	Port         string `json:"port,omitempty"`
	IdentityFile string `json:"identity_file,omitempty"`
	ProxyJump    string `json:"proxy_jump,omitempty"`
	ForwardAgent bool   `json:"forward_agent,omitempty"`
}

// SSHKeyRef is a reference to an SSH key (no private data).
type SSHKeyRef struct {
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
	Path string `json:"path"`
}

// GitExport holds portable git configuration.
type GitExport struct {
	UserName  string             `json:"user_name,omitempty"`
	UserEmail string             `json:"user_email,omitempty"`
	Profiles  []GitProfileExport `json:"profiles,omitempty"`
}

// GitProfileExport represents a git includeIf profile for migration.
type GitProfileExport struct {
	Name      string `json:"name"`
	Directory string `json:"directory"`
	UserName  string `json:"user_name,omitempty"`
	Email     string `json:"email,omitempty"`
	SignKey   string `json:"sign_key,omitempty"`
}
