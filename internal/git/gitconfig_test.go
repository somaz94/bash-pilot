package git

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTestFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseGitConfigFile(t *testing.T) {
	dir := t.TempDir()
	cfg := writeTestFile(t, dir, ".gitconfig", `[user]
	name = Test User
	email = test@example.com
[safe]
	directory = /home/user/repo1
	directory = /home/user/repo2
[remote "origin"]
	url = git@github.com:user/repo.git
	fetch = +refs/heads/*:refs/remotes/origin/*
`)

	sections, err := ParseGitConfigFile(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sections) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(sections))
	}

	if sections[0].Header != "user" {
		t.Errorf("expected section 'user', got %q", sections[0].Header)
	}
	if len(sections[0].Entries) != 2 {
		t.Errorf("expected 2 entries in user section, got %d", len(sections[0].Entries))
	}

	if sections[1].Header != "safe" {
		t.Errorf("expected section 'safe', got %q", sections[1].Header)
	}
	if len(sections[1].Entries) != 2 {
		t.Errorf("expected 2 entries in safe section, got %d", len(sections[1].Entries))
	}
}

func TestParseGitConfigFile_Comments(t *testing.T) {
	dir := t.TempDir()
	cfg := writeTestFile(t, dir, ".gitconfig", `# This is a comment
; This is also a comment
[user]
	email = test@example.com
`)

	sections, err := ParseGitConfigFile(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}
}

func TestParseGitConfigFile_NotFound(t *testing.T) {
	_, err := ParseGitConfigFile("/nonexistent/file")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestGetProfiles_IncludeIf(t *testing.T) {
	dir := t.TempDir()

	// Create included config.
	writeTestFile(t, dir, ".gitconfig-work", `[user]
	email = work@company.com
	signingkey = ABC123
`)

	// Create main config with includeIf.
	cfg := writeTestFile(t, dir, ".gitconfig", `[user]
	name = Global User
	email = global@example.com
[includeIf "gitdir:`+dir+`/work/"]
	path = `+dir+`/.gitconfig-work
`)

	profiles, err := GetProfiles(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(profiles) < 2 {
		t.Fatalf("expected at least 2 profiles, got %d", len(profiles))
	}

	// First should be global.
	if profiles[0].Email != "global@example.com" {
		t.Errorf("expected global email, got %q", profiles[0].Email)
	}

	// Second should be work profile.
	if profiles[1].Email != "work@company.com" {
		t.Errorf("expected work email, got %q", profiles[1].Email)
	}
	if profiles[1].SignKey != "ABC123" {
		t.Errorf("expected sign key ABC123, got %q", profiles[1].SignKey)
	}
}

func TestGetProfiles_IncludeIF_CaseInsensitive(t *testing.T) {
	dir := t.TempDir()

	writeTestFile(t, dir, ".gitconfig-work", `[user]
	email = work@company.com
`)

	// Use includeIF (uppercase IF) like real-world gitconfig.
	cfg := writeTestFile(t, dir, ".gitconfig", `[user]
	name = Test User
	email = global@example.com
[includeIF "gitdir:`+dir+`/work/"]
	path = `+dir+`/.gitconfig-work
`)

	profiles, err := GetProfiles(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(profiles) < 2 {
		t.Fatalf("expected at least 2 profiles, got %d", len(profiles))
	}

	found := false
	for _, p := range profiles {
		if p.Email == "work@company.com" {
			found = true
		}
	}
	if !found {
		t.Error("includeIF (uppercase) should be recognized as includeIf")
	}
}

func TestGetProfiles_GlobalOnly(t *testing.T) {
	dir := t.TempDir()
	cfg := writeTestFile(t, dir, ".gitconfig", `[user]
	name = Solo User
	email = solo@example.com
`)

	profiles, err := GetProfiles(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}

	if !profiles[0].Active {
		t.Error("global profile should be active when no other profiles exist")
	}
}

func TestGetProfiles_NoEmail(t *testing.T) {
	dir := t.TempDir()
	cfg := writeTestFile(t, dir, ".gitconfig", `[core]
	autocrlf = input
`)

	profiles, err := GetProfiles(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(profiles) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(profiles))
	}
}

func TestDoctor_DuplicateSafeDirs(t *testing.T) {
	dir := t.TempDir()
	cfg := writeTestFile(t, dir, ".gitconfig", `[safe]
	directory = /home/user/repo1
	directory = /home/user/repo1
	directory = /home/user/repo2
`)

	// Set restrictive permissions to avoid permission warning.
	os.Chmod(cfg, 0600)

	result, err := Doctor(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, issue := range result.Issues {
		if issue.Category == "safe.directory" && issue.Severity == "warn" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected warning about duplicate safe.directory")
	}
}

func TestDoctor_MissingIncludeIfTarget(t *testing.T) {
	dir := t.TempDir()
	cfg := writeTestFile(t, dir, ".gitconfig", `[includeIf "gitdir:~/work/"]
	path = /nonexistent/path/.gitconfig-work
`)

	os.Chmod(cfg, 0600)

	result, err := Doctor(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, issue := range result.Issues {
		if issue.Category == "includeIf" && issue.Severity == "warn" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected warning about missing includeIf target")
	}
}

func TestDoctor_NoUserEmail(t *testing.T) {
	dir := t.TempDir()
	cfg := writeTestFile(t, dir, ".gitconfig", `[core]
	autocrlf = input
`)

	os.Chmod(cfg, 0600)

	result, err := Doctor(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, issue := range result.Issues {
		if issue.Category == "user.identity" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected warning about missing user.email")
	}
}

func TestDoctor_NoUserEmail_WithIncludeIf(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, ".gitconfig-work", `[user]
	email = work@example.com
`)
	cfg := writeTestFile(t, dir, ".gitconfig", `[includeIf "gitdir:~/work/"]
	path = `+dir+`/.gitconfig-work
`)

	os.Chmod(cfg, 0600)

	result, err := Doctor(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, issue := range result.Issues {
		if issue.Category == "user.identity" && issue.Severity == "warn" {
			t.Error("should not warn about missing email when includeIf provides it")
		}
	}
}

func TestDoctor_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	cfg := writeTestFile(t, dir, ".gitconfig", `[user]
	email = test@example.com
`)

	// Set overly permissive.
	os.Chmod(cfg, 0644)

	result, err := Doctor(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, issue := range result.Issues {
		if issue.Category == "permissions" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected warning about file permissions")
	}
}

func TestDoctor_Clean(t *testing.T) {
	dir := t.TempDir()
	cfg := writeTestFile(t, dir, ".gitconfig", `[user]
	email = test@example.com
`)

	os.Chmod(cfg, 0600)

	result, err := Doctor(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have no issues (only ok).
	for _, issue := range result.Issues {
		if issue.Severity != "ok" {
			t.Errorf("unexpected issue: %v", issue)
		}
	}
}

func TestDoctor_NonexistentFile(t *testing.T) {
	result, err := Doctor("/nonexistent/file")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Issues) == 0 {
		t.Fatal("expected at least one issue")
	}
	if result.Issues[0].Severity != "error" {
		t.Errorf("expected error severity, got %q", result.Issues[0].Severity)
	}
}

func TestFindDuplicateSafeDirs(t *testing.T) {
	sections := []Section{
		{
			Header: "safe",
			Entries: []Entry{
				{Key: "directory", Value: "/repo1", Line: 2},
				{Key: "directory", Value: "/repo2", Line: 3},
				{Key: "directory", Value: "/repo1", Line: 4},
				{Key: "directory", Value: "/repo1", Line: 5},
			},
		},
	}

	dups := FindDuplicateSafeDirs(sections)
	if len(dups) != 2 {
		t.Fatalf("expected 2 duplicates, got %d", len(dups))
	}
}

func TestClean_DryRun(t *testing.T) {
	dir := t.TempDir()
	cfg := writeTestFile(t, dir, ".gitconfig", `[safe]
	directory = /repo1
	directory = /repo1
	directory = /repo2
`)

	result, err := Clean(cfg, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.DryRun {
		t.Error("expected dry run mode")
	}
	if len(result.Removed) == 0 {
		t.Error("expected entries to be marked for removal")
	}

	// File should not be modified.
	data, _ := os.ReadFile(cfg)
	if len(data) == 0 {
		t.Error("file should not be empty")
	}
}

func TestClean_Actual(t *testing.T) {
	dir := t.TempDir()
	cfg := writeTestFile(t, dir, ".gitconfig", `[user]
	email = test@example.com
[safe]
	directory = /repo1
	directory = /repo1
	directory = /repo2
`)

	result, err := Clean(cfg, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.DryRun {
		t.Error("should not be dry run")
	}
	if len(result.Removed) == 0 {
		t.Error("expected entries to be removed")
	}
	if result.BackupDir == "" {
		t.Error("expected backup path")
	}

	// Verify backup exists.
	if _, err := os.Stat(result.BackupDir); os.IsNotExist(err) {
		t.Error("backup file should exist")
	}

	// Verify file was modified - should have fewer lines.
	data, _ := os.ReadFile(cfg)
	content := string(data)
	if content == "" {
		t.Error("file should not be empty after clean")
	}
}

func TestClean_NoDuplicates(t *testing.T) {
	dir := t.TempDir()
	// Use directories that actually exist to avoid stale removal.
	cfg := writeTestFile(t, dir, ".gitconfig", `[safe]
	directory = `+dir+`
	directory = /tmp
`)

	result, err := Clean(cfg, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Removed) != 0 {
		t.Errorf("expected no removals, got %d", len(result.Removed))
	}
}

func TestClean_NonexistentFile(t *testing.T) {
	_, err := Clean("/nonexistent/file", false)
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input    string
		expected string
	}{
		{"~/test", filepath.Join(home, "test")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		result := expandPath(tt.input)
		if result != tt.expected {
			t.Errorf("expandPath(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestRemoveEmptySections(t *testing.T) {
	input := "[safe]\n\n[user]\n\temail = test@example.com\n"
	result := removeEmptySections(input, "safe")

	if result == input {
		t.Error("empty [safe] section should be removed")
	}
}

func TestDefaultGitConfigPath(t *testing.T) {
	path := DefaultGitConfigPath()
	if path == "" {
		t.Error("should return a non-empty path")
	}
}

func TestDoctor_DuplicateRemotes(t *testing.T) {
	dir := t.TempDir()
	cfg := writeTestFile(t, dir, ".gitconfig", `[remote "origin"]
	url = git@github.com:user/repo.git
[remote "backup"]
	url = git@github.com:user/repo.git
`)

	os.Chmod(cfg, 0600)

	result, err := Doctor(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, issue := range result.Issues {
		if issue.Category == "remote" && issue.Severity == "warn" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected warning about duplicate remote URLs")
	}
}

func TestGetProfiles_MissingIncludePath(t *testing.T) {
	dir := t.TempDir()
	cfg := writeTestFile(t, dir, ".gitconfig", `[includeIf "gitdir:~/work/"]
	somekey = somevalue
`)

	profiles, err := GetProfiles(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No profiles expected since no path entry in includeIf.
	for _, p := range profiles {
		if p.Directory == "~/work" {
			t.Error("should not create profile without path entry")
		}
	}
}

func TestGetProfiles_NonGitdirCondition(t *testing.T) {
	dir := t.TempDir()
	cfg := writeTestFile(t, dir, ".gitconfig", `[user]
	email = test@example.com
[includeIf "onbranch:main"]
	path = /some/path
`)

	profiles, err := GetProfiles(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only have global profile, not onbranch.
	if len(profiles) != 1 {
		t.Errorf("expected 1 profile (global only), got %d", len(profiles))
	}
}

func TestGetProfiles_ActiveProfile(t *testing.T) {
	dir := t.TempDir()

	// Create work config.
	writeTestFile(t, dir, ".gitconfig-work", `[user]
	email = work@company.com
`)

	// Use current working directory as the gitdir match.
	cwd, _ := os.Getwd()

	cfg := writeTestFile(t, dir, ".gitconfig", `[user]
	email = global@example.com
[includeIf "gitdir:`+cwd+`/"]
	path = `+dir+`/.gitconfig-work
`)

	profiles, err := GetProfiles(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Work profile should be active.
	workActive := false
	globalActive := false
	for _, p := range profiles {
		if p.Email == "work@company.com" && p.Active {
			workActive = true
		}
		if p.Email == "global@example.com" && p.Active {
			globalActive = true
		}
	}

	if !workActive {
		t.Error("work profile should be active (cwd matches)")
	}
	if globalActive {
		t.Error("global should not be active when another profile is active")
	}
}

func TestGetProfiles_ActiveProfileNoPrefixFalsePositive(t *testing.T) {
	dir := t.TempDir()

	writeTestFile(t, dir, ".gitconfig-work", `[user]
	email = work@company.com
`)

	cwd, _ := os.Getwd()

	// gitdir is cwd+"-backup" — a different directory that starts with cwd as prefix.
	// The work profile should NOT be active.
	cfg := writeTestFile(t, dir, ".gitconfig", `[user]
	email = global@example.com
[includeIf "gitdir:`+cwd+`-backup/"]
	path = `+dir+`/.gitconfig-work
`)

	profiles, err := GetProfiles(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, p := range profiles {
		if p.Email == "work@company.com" && p.Active {
			t.Errorf("work profile should NOT be active: cwd=%s, gitdir=%s-backup", cwd, cwd)
		}
	}
}

func TestFindDuplicateSafeDirs_NoDuplicates(t *testing.T) {
	sections := []Section{
		{
			Header: "safe",
			Entries: []Entry{
				{Key: "directory", Value: "/repo1", Line: 2},
				{Key: "directory", Value: "/repo2", Line: 3},
			},
		},
	}

	dups := FindDuplicateSafeDirs(sections)
	if len(dups) != 0 {
		t.Errorf("expected 0 duplicates, got %d", len(dups))
	}
}

func TestFindDuplicateSafeDirs_NonSafeSection(t *testing.T) {
	sections := []Section{
		{
			Header: "user",
			Entries: []Entry{
				{Key: "directory", Value: "/repo1", Line: 2},
			},
		},
	}

	dups := FindDuplicateSafeDirs(sections)
	if len(dups) != 0 {
		t.Errorf("expected 0 duplicates, got %d", len(dups))
	}
}

func TestClean_StaleDirectories(t *testing.T) {
	dir := t.TempDir()
	cfg := writeTestFile(t, dir, ".gitconfig", `[safe]
	directory = /nonexistent/path/that/does/not/exist
	directory = /another/nonexistent/path
`)

	result, err := Clean(cfg, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Removed) == 0 {
		t.Error("expected stale directories to be removed")
	}
}

func TestCheckIncludeIfs_ValidTarget(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, ".gitconfig-work", `[user]
	email = work@example.com
`)
	cfg := writeTestFile(t, dir, ".gitconfig", `[includeIf "gitdir:~/work/"]
	path = `+dir+`/.gitconfig-work
`)

	os.Chmod(cfg, 0600)

	result, err := Doctor(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, issue := range result.Issues {
		if issue.Category == "includeIf" && issue.Severity == "warn" {
			t.Error("should not warn about valid includeIf target")
		}
	}
}

func TestGetGitConfigValue(t *testing.T) {
	sections := []Section{
		{
			Header: "user",
			Entries: []Entry{
				{Key: "name", Value: "Test"},
				{Key: "email", Value: "test@example.com"},
			},
		},
	}

	email := getGitConfigValue(sections, "user", "email")
	if email != "test@example.com" {
		t.Errorf("expected test@example.com, got %q", email)
	}

	missing := getGitConfigValue(sections, "user", "nonexistent")
	if missing != "" {
		t.Errorf("expected empty string, got %q", missing)
	}

	noSection := getGitConfigValue(sections, "nonexistent", "email")
	if noSection != "" {
		t.Errorf("expected empty string, got %q", noSection)
	}
}
