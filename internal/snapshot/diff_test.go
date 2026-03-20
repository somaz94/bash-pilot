package snapshot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompareField(t *testing.T) {
	tests := []struct {
		name             string
		key, left, right string
		expectedStatus   string
	}{
		{"match", "OS", "darwin", "darwin", "match"},
		{"mismatch", "OS", "darwin", "linux", "mismatch"},
		{"missing", "Editor", "vim", "", "missing"},
		{"extra", "Editor", "", "vim", "extra"},
		{"both empty", "Editor", "", "", "match"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := compareField(tt.key, tt.left, tt.right)
			if entry.Status != tt.expectedStatus {
				t.Errorf("expected status %s, got %s", tt.expectedStatus, entry.Status)
			}
			if entry.Key != tt.key {
				t.Errorf("expected key %s, got %s", tt.key, entry.Key)
			}
		})
	}
}

func TestDiff_System(t *testing.T) {
	origRunCommand := runCommand
	origLookPath := lookPath
	origGetenv := getenv
	origUserHomeDir := userHomeDir
	defer func() {
		runCommand = origRunCommand
		lookPath = origLookPath
		getenv = origGetenv
		userHomeDir = origUserHomeDir
	}()

	// Mock everything to produce a known current snapshot.
	lookPath = func(file string) (string, error) {
		return "", errNotFound
	}
	runCommand = func(name string, args ...string) ([]byte, error) {
		return nil, errNotFound
	}
	getenv = func(key string) string {
		switch key {
		case "SHELL":
			return "/bin/zsh"
		case "EDITOR":
			return "nano"
		}
		return ""
	}
	userHomeDir = func() (string, error) {
		return "/tmp/nonexistent-home-for-test", nil
	}

	saved := &Snapshot{
		OS:   "darwin",
		Arch: "arm64",
		Shell: ShellInfo{
			Shell:  "/bin/bash",
			Editor: "vim",
		},
	}

	result := Diff(saved)

	// Find System section.
	var sys *DiffSection
	for i := range result.Sections {
		if result.Sections[i].Name == "System" {
			sys = &result.Sections[i]
			break
		}
	}
	if sys == nil {
		t.Fatal("System section not found")
	}

	// Check Shell field shows mismatch.
	shellFound := false
	for _, e := range sys.Entries {
		if e.Key == "Shell" && e.Status == "mismatch" {
			shellFound = true
			if e.Left != "/bin/bash" || e.Right != "/bin/zsh" {
				t.Errorf("expected shell mismatch /bin/bash → /bin/zsh, got %s → %s", e.Left, e.Right)
			}
		}
	}
	if !shellFound {
		t.Error("expected Shell mismatch entry")
	}
}

var errNotFound = &notFoundError{}

type notFoundError struct{}

func (e *notFoundError) Error() string { return "not found" }

func TestDiff_Tools(t *testing.T) {
	origRunCommand := runCommand
	origLookPath := lookPath
	origGetenv := getenv
	origUserHomeDir := userHomeDir
	defer func() {
		runCommand = origRunCommand
		lookPath = origLookPath
		getenv = origGetenv
		userHomeDir = origUserHomeDir
	}()

	lookPath = func(file string) (string, error) {
		if file == "git" {
			return "/usr/bin/git", nil
		}
		if file == "node" {
			return "/usr/local/bin/node", nil
		}
		return "", errNotFound
	}
	runCommand = func(name string, args ...string) ([]byte, error) {
		if name == "git" {
			return []byte("git version 2.43.0"), nil
		}
		if name == "node" {
			return []byte("v21.0.0"), nil
		}
		return nil, errNotFound
	}
	getenv = func(key string) string { return "" }
	userHomeDir = func() (string, error) { return "/tmp/nonexistent-home-for-test", nil }

	saved := &Snapshot{
		Tools: []ToolInfo{
			{Name: "git", Version: "git version 2.42.0"}, // version mismatch
			{Name: "curl", Version: "curl 8.0.0"},        // missing in current
		},
	}

	result := Diff(saved)

	var tools *DiffSection
	for i := range result.Sections {
		if result.Sections[i].Name == "Tools" {
			tools = &result.Sections[i]
			break
		}
	}
	if tools == nil {
		t.Fatal("Tools section not found")
	}

	entryMap := make(map[string]DiffEntry)
	for _, e := range tools.Entries {
		entryMap[e.Key] = e
	}

	if e, ok := entryMap["git"]; !ok || e.Status != "mismatch" {
		t.Error("expected git mismatch")
	}
	if e, ok := entryMap["curl"]; !ok || e.Status != "missing" {
		t.Error("expected curl missing")
	}
	if e, ok := entryMap["node"]; !ok || e.Status != "extra" {
		t.Error("expected node extra")
	}
}

func TestDiff_SSHKeys(t *testing.T) {
	origRunCommand := runCommand
	origLookPath := lookPath
	origGetenv := getenv
	origUserHomeDir := userHomeDir
	defer func() {
		runCommand = origRunCommand
		lookPath = origLookPath
		getenv = origGetenv
		userHomeDir = origUserHomeDir
	}()

	lookPath = func(file string) (string, error) { return "", errNotFound }
	runCommand = func(name string, args ...string) ([]byte, error) { return nil, errNotFound }
	getenv = func(key string) string { return "" }
	userHomeDir = func() (string, error) { return "/tmp/nonexistent-home-for-test", nil }

	saved := &Snapshot{
		SSHKeys: []SSHKeyInfo{
			{Name: "id_ed25519", Fingerprint: "SHA256:old"},
			{Name: "id_rsa", Fingerprint: "SHA256:rsa"},
		},
	}

	result := Diff(saved)

	var sshSection *DiffSection
	for i := range result.Sections {
		if result.Sections[i].Name == "SSH Keys" {
			sshSection = &result.Sections[i]
			break
		}
	}
	if sshSection == nil {
		t.Fatal("SSH Keys section not found")
	}

	// Both should be missing since current has no SSH keys.
	for _, e := range sshSection.Entries {
		if e.Status != "missing" {
			t.Errorf("expected all SSH keys missing, got %s for %s", e.Status, e.Key)
		}
	}
}

func TestDiff_K8sContexts(t *testing.T) {
	origRunCommand := runCommand
	origLookPath := lookPath
	origGetenv := getenv
	origUserHomeDir := userHomeDir
	defer func() {
		runCommand = origRunCommand
		lookPath = origLookPath
		getenv = origGetenv
		userHomeDir = origUserHomeDir
	}()

	lookPath = func(file string) (string, error) { return "", errNotFound }
	runCommand = func(name string, args ...string) ([]byte, error) { return nil, errNotFound }
	getenv = func(key string) string { return "" }
	userHomeDir = func() (string, error) { return "/tmp/nonexistent-home-for-test", nil }

	saved := &Snapshot{
		K8s: []K8sContext{
			{Name: "prod-cluster", Current: true},
		},
	}

	result := Diff(saved)

	var k8s *DiffSection
	for i := range result.Sections {
		if result.Sections[i].Name == "K8s Contexts" {
			k8s = &result.Sections[i]
			break
		}
	}
	if k8s == nil {
		t.Fatal("K8s Contexts section not found")
	}
	if len(k8s.Entries) != 1 || k8s.Entries[0].Status != "missing" {
		t.Error("expected prod-cluster to be missing")
	}
}

func TestDiff_BrewPackages(t *testing.T) {
	origRunCommand := runCommand
	origLookPath := lookPath
	origGetenv := getenv
	origUserHomeDir := userHomeDir
	defer func() {
		runCommand = origRunCommand
		lookPath = origLookPath
		getenv = origGetenv
		userHomeDir = origUserHomeDir
	}()

	lookPath = func(file string) (string, error) { return "", errNotFound }
	runCommand = func(name string, args ...string) ([]byte, error) { return nil, errNotFound }
	getenv = func(key string) string { return "" }
	userHomeDir = func() (string, error) { return "/tmp/nonexistent-home-for-test", nil }

	saved := &Snapshot{
		Brew: []string{"git", "wget"},
	}

	result := Diff(saved)

	var brew *DiffSection
	for i := range result.Sections {
		if result.Sections[i].Name == "Brew Packages" {
			brew = &result.Sections[i]
			break
		}
	}
	if brew == nil {
		t.Fatal("Brew Packages section not found")
	}
	for _, e := range brew.Entries {
		if e.Status != "missing" {
			t.Errorf("expected all brew packages missing, got %s for %s", e.Status, e.Key)
		}
	}
}

func TestDiff_Summary(t *testing.T) {
	origRunCommand := runCommand
	origLookPath := lookPath
	origGetenv := getenv
	origUserHomeDir := userHomeDir
	defer func() {
		runCommand = origRunCommand
		lookPath = origLookPath
		getenv = origGetenv
		userHomeDir = origUserHomeDir
	}()

	lookPath = func(file string) (string, error) { return "", errNotFound }
	runCommand = func(name string, args ...string) ([]byte, error) { return nil, errNotFound }
	getenv = func(key string) string {
		if key == "SHELL" {
			return "/bin/zsh"
		}
		return ""
	}
	userHomeDir = func() (string, error) { return "/tmp/nonexistent-home-for-test", nil }

	saved := &Snapshot{
		OS:    "darwin",
		Shell: ShellInfo{Shell: "/bin/zsh"},
	}

	result := Diff(saved)

	if result.Summary.Match < 1 {
		t.Error("expected at least 1 match in summary")
	}
}

func TestFormatDiff(t *testing.T) {
	result := &DiffResult{
		Sections: []DiffSection{
			{
				Name: "System",
				Entries: []DiffEntry{
					{Status: "match", Key: "OS", Left: "darwin", Right: "darwin"},
					{Status: "match", Key: "Arch", Left: "arm64", Right: "arm64"},
				},
			},
			{
				Name: "Tools",
				Entries: []DiffEntry{
					{Status: "mismatch", Key: "git", Left: "2.42.0", Right: "2.43.0"},
					{Status: "missing", Key: "curl", Left: "8.0.0"},
					{Status: "extra", Key: "node", Right: "v21.0.0"},
				},
			},
		},
		Summary: DiffSummary{Match: 2, Mismatch: 1, Missing: 1, Extra: 1},
	}

	output := FormatDiff(result)

	if !strings.Contains(output, "all match") {
		t.Error("expected 'all match' for System section")
	}
	if !strings.Contains(output, "~ git") {
		t.Error("expected mismatch indicator for git")
	}
	if !strings.Contains(output, "- curl") {
		t.Error("expected missing indicator for curl")
	}
	if !strings.Contains(output, "+ node") {
		t.Error("expected extra indicator for node")
	}
	if !strings.Contains(output, "2 match") {
		t.Error("expected summary counts")
	}
}

func TestDiff_ToolsMatch(t *testing.T) {
	origRunCommand := runCommand
	origLookPath := lookPath
	origGetenv := getenv
	origUserHomeDir := userHomeDir
	defer func() {
		runCommand = origRunCommand
		lookPath = origLookPath
		getenv = origGetenv
		userHomeDir = origUserHomeDir
	}()

	lookPath = func(file string) (string, error) {
		if file == "git" {
			return "/usr/bin/git", nil
		}
		return "", errNotFound
	}
	runCommand = func(name string, args ...string) ([]byte, error) {
		if name == "git" && len(args) > 0 && args[0] == "--version" {
			return []byte("git version 2.42.0"), nil
		}
		return nil, errNotFound
	}
	getenv = func(key string) string { return "" }
	userHomeDir = func() (string, error) { return "/tmp/nonexistent-home-for-test", nil }

	saved := &Snapshot{
		Tools: []ToolInfo{
			{Name: "git", Version: "git version 2.42.0"},
		},
	}

	result := Diff(saved)

	var tools *DiffSection
	for i := range result.Sections {
		if result.Sections[i].Name == "Tools" {
			tools = &result.Sections[i]
			break
		}
	}
	if tools == nil {
		t.Fatal("Tools section not found")
	}

	for _, e := range tools.Entries {
		if e.Key == "git" && e.Status != "match" {
			t.Errorf("expected git match, got %s", e.Status)
		}
	}
}

func TestDiff_SSHKeysMatch(t *testing.T) {
	origRunCommand := runCommand
	origLookPath := lookPath
	origGetenv := getenv
	origUserHomeDir := userHomeDir
	defer func() {
		runCommand = origRunCommand
		lookPath = origLookPath
		getenv = origGetenv
		userHomeDir = origUserHomeDir
	}()

	lookPath = func(file string) (string, error) { return "", errNotFound }
	getenv = func(key string) string { return "" }

	tmpDir := t.TempDir()
	userHomeDir = func() (string, error) { return tmpDir, nil }

	// Create SSH key file.
	sshDir := filepath.Join(tmpDir, ".ssh")
	os.MkdirAll(sshDir, 0700)
	os.WriteFile(filepath.Join(sshDir, "id_ed25519"), []byte("key"), 0600)

	runCommand = func(name string, args ...string) ([]byte, error) {
		if name == "ssh-keygen" {
			return []byte("256 SHA256:abc123 user@host (ED25519)\n"), nil
		}
		return nil, errNotFound
	}

	saved := &Snapshot{
		SSHKeys: []SSHKeyInfo{
			{Name: "id_ed25519", Fingerprint: "SHA256:abc123"},
		},
	}

	result := Diff(saved)

	var sshSection *DiffSection
	for i := range result.Sections {
		if result.Sections[i].Name == "SSH Keys" {
			sshSection = &result.Sections[i]
			break
		}
	}
	if sshSection == nil {
		t.Fatal("SSH Keys section not found")
	}

	found := false
	for _, e := range sshSection.Entries {
		if e.Key == "id_ed25519" && e.Status == "match" {
			found = true
		}
	}
	if !found {
		t.Error("expected id_ed25519 to match")
	}
}

func TestDiff_SSHKeysMismatch(t *testing.T) {
	origRunCommand := runCommand
	origLookPath := lookPath
	origGetenv := getenv
	origUserHomeDir := userHomeDir
	defer func() {
		runCommand = origRunCommand
		lookPath = origLookPath
		getenv = origGetenv
		userHomeDir = origUserHomeDir
	}()

	lookPath = func(file string) (string, error) { return "", errNotFound }
	getenv = func(key string) string { return "" }

	tmpDir := t.TempDir()
	userHomeDir = func() (string, error) { return tmpDir, nil }

	sshDir := filepath.Join(tmpDir, ".ssh")
	os.MkdirAll(sshDir, 0700)
	os.WriteFile(filepath.Join(sshDir, "id_ed25519"), []byte("key"), 0600)

	runCommand = func(name string, args ...string) ([]byte, error) {
		if name == "ssh-keygen" {
			return []byte("256 SHA256:new456 user@host (ED25519)\n"), nil
		}
		return nil, errNotFound
	}

	saved := &Snapshot{
		SSHKeys: []SSHKeyInfo{
			{Name: "id_ed25519", Fingerprint: "SHA256:old123"},
		},
	}

	result := Diff(saved)

	var sshSection *DiffSection
	for i := range result.Sections {
		if result.Sections[i].Name == "SSH Keys" {
			sshSection = &result.Sections[i]
			break
		}
	}
	if sshSection == nil {
		t.Fatal("SSH Keys section not found")
	}

	found := false
	for _, e := range sshSection.Entries {
		if e.Key == "id_ed25519" && e.Status == "mismatch" {
			found = true
		}
	}
	if !found {
		t.Error("expected id_ed25519 fingerprint mismatch")
	}
}

func TestDiff_K8sMatch(t *testing.T) {
	origRunCommand := runCommand
	origLookPath := lookPath
	origGetenv := getenv
	origUserHomeDir := userHomeDir
	defer func() {
		runCommand = origRunCommand
		lookPath = origLookPath
		getenv = origGetenv
		userHomeDir = origUserHomeDir
	}()

	lookPath = func(file string) (string, error) {
		if file == "kubectl" {
			return "/usr/local/bin/kubectl", nil
		}
		return "", errNotFound
	}
	runCommand = func(name string, args ...string) ([]byte, error) {
		if name == "kubectl" {
			joined := strings.Join(args, " ")
			if strings.Contains(joined, "current-context") {
				return []byte("cluster-a\n"), nil
			}
			if strings.Contains(joined, "get-contexts") {
				return []byte("cluster-a\ncluster-b\n"), nil
			}
		}
		return nil, errNotFound
	}
	getenv = func(key string) string { return "" }
	userHomeDir = func() (string, error) { return "/tmp/nonexistent-home-for-test", nil }

	saved := &Snapshot{
		K8s: []K8sContext{
			{Name: "cluster-a"},
			{Name: "cluster-c"}, // missing in current
		},
	}

	result := Diff(saved)

	var k8s *DiffSection
	for i := range result.Sections {
		if result.Sections[i].Name == "K8s Contexts" {
			k8s = &result.Sections[i]
			break
		}
	}
	if k8s == nil {
		t.Fatal("K8s section not found")
	}

	entryMap := make(map[string]string)
	for _, e := range k8s.Entries {
		entryMap[e.Key] = e.Status
	}

	if entryMap["cluster-a"] != "match" {
		t.Error("expected cluster-a match")
	}
	if entryMap["cluster-c"] != "missing" {
		t.Error("expected cluster-c missing")
	}
	if entryMap["cluster-b"] != "extra" {
		t.Error("expected cluster-b extra")
	}
}

func TestDiff_BrewMatch(t *testing.T) {
	origRunCommand := runCommand
	origLookPath := lookPath
	origGetenv := getenv
	origUserHomeDir := userHomeDir
	defer func() {
		runCommand = origRunCommand
		lookPath = origLookPath
		getenv = origGetenv
		userHomeDir = origUserHomeDir
	}()

	lookPath = func(file string) (string, error) { return "", errNotFound }
	runCommand = func(name string, args ...string) ([]byte, error) { return nil, errNotFound }
	getenv = func(key string) string { return "" }
	userHomeDir = func() (string, error) { return "/tmp/nonexistent-home-for-test", nil }

	// Both saved and current empty — section should be skipped.
	saved := &Snapshot{}
	result := Diff(saved)

	for _, sec := range result.Sections {
		if sec.Name == "Brew Packages" {
			t.Error("did not expect Brew Packages section when both are empty")
		}
	}
}

func TestFormatDiff_AllMatch(t *testing.T) {
	result := &DiffResult{
		Sections: []DiffSection{
			{
				Name: "System",
				Entries: []DiffEntry{
					{Status: "match", Key: "OS"},
				},
			},
		},
		Summary: DiffSummary{Match: 1},
	}

	output := FormatDiff(result)
	if !strings.Contains(output, "all match") {
		t.Error("expected all match output")
	}
}

func TestParseOnly(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]bool
	}{
		{"empty", "", nil},
		{"single", "ssh", map[string]bool{"ssh": true}},
		{"multiple", "ssh,git", map[string]bool{"ssh": true, "git": true}},
		{"with spaces", " ssh , git ", map[string]bool{"ssh": true, "git": true}},
		{"uppercase", "SSH,Git", map[string]bool{"ssh": true, "git": true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseOnly(tt.input)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d entries, got %d", len(tt.expected), len(result))
			}
			for k := range tt.expected {
				if !result[k] {
					t.Errorf("expected key %s", k)
				}
			}
		})
	}
}

func TestDiff_OnlyFilter(t *testing.T) {
	origRunCommand := runCommand
	origLookPath := lookPath
	origGetenv := getenv
	origUserHomeDir := userHomeDir
	defer func() {
		runCommand = origRunCommand
		lookPath = origLookPath
		getenv = origGetenv
		userHomeDir = origUserHomeDir
	}()

	lookPath = func(file string) (string, error) { return "", errNotFound }
	runCommand = func(name string, args ...string) ([]byte, error) { return nil, errNotFound }
	getenv = func(key string) string {
		if key == "SHELL" {
			return "/bin/zsh"
		}
		return ""
	}
	userHomeDir = func() (string, error) { return "/tmp/nonexistent-home-for-test", nil }

	saved := &Snapshot{
		OS:   "darwin",
		Arch: "arm64",
		Shell: ShellInfo{
			Shell: "/bin/zsh",
		},
		Tools: []ToolInfo{
			{Name: "git", Version: "git version 2.42.0"},
		},
		Brew: []string{"wget"},
	}

	// Only tools
	only := map[string]bool{"tools": true}
	result := Diff(saved, only)

	sectionNames := make(map[string]bool)
	for _, sec := range result.Sections {
		sectionNames[sec.Name] = true
	}

	if !sectionNames["Tools"] {
		t.Error("expected Tools section")
	}
	if sectionNames["System"] {
		t.Error("did not expect System section with --only tools")
	}
	if sectionNames["Git"] {
		t.Error("did not expect Git section with --only tools")
	}
	if sectionNames["SSH Keys"] {
		t.Error("did not expect SSH Keys section with --only tools")
	}
	if sectionNames["Brew Packages"] {
		t.Error("did not expect Brew Packages section with --only tools")
	}

	// Only ssh,git
	only2 := map[string]bool{"ssh": true, "git": true}
	result2 := Diff(saved, only2)

	sectionNames2 := make(map[string]bool)
	for _, sec := range result2.Sections {
		sectionNames2[sec.Name] = true
	}

	if !sectionNames2["Git"] {
		t.Error("expected Git section")
	}
	if !sectionNames2["SSH Keys"] {
		t.Error("expected SSH Keys section")
	}
	if sectionNames2["System"] {
		t.Error("did not expect System section")
	}
	if sectionNames2["Tools"] {
		t.Error("did not expect Tools section")
	}
}
