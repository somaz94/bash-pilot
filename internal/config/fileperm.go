package config

import "os"

// File modes the tool creates files/dirs with; callers name the intent, not an octal literal.
const (
	// PermSSHDir is the mode for the user's ~/.ssh directory (owner-only).
	PermSSHDir os.FileMode = 0700

	// PermSSHConfigFile is the mode for ~/.ssh/config (owner-only read/write).
	PermSSHConfigFile os.FileMode = 0600

	// PermGitConfigFile is the mode for ~/.gitconfig, included gitconfig
	// profiles, and gitconfig backup files.
	PermGitConfigFile os.FileMode = 0600

	// PermConfigDir is the mode for non-secret directories the tool creates:
	// its own config directory and git profile working directories.
	PermConfigDir os.FileMode = 0755

	// PermConfigFile is the mode for bash-pilot's main config file and other
	// non-secret config files created by the tool.
	PermConfigFile os.FileMode = 0644
)
