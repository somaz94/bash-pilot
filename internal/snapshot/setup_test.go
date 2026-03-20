package snapshot

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func TestPlan_MissingTools(t *testing.T) {
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

	// Current env has no tools.
	lookPath = func(file string) (string, error) {
		return "", fmt.Errorf("not found")
	}
	runCommand = func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("not found")
	}
	getenv = func(key string) string { return "" }
	userHomeDir = func() (string, error) { return "/tmp/nonexistent-home-for-test", nil }

	saved := &Snapshot{
		Tools: []ToolInfo{
			{Name: "git", Version: "git version 2.42.0"},
			{Name: "curl", Version: "curl 8.0.0"},
		},
	}

	result := Plan(saved)

	if len(result.Actions) < 2 {
		t.Fatalf("expected at least 2 actions, got %d", len(result.Actions))
	}

	actionMap := make(map[string]SetupAction)
	for _, a := range result.Actions {
		actionMap[a.Tool] = a
	}

	gitAction, ok := actionMap["git"]
	if !ok {
		t.Fatal("expected git action")
	}
	if gitAction.Status != "pending" {
		t.Errorf("expected git status pending, got %s", gitAction.Status)
	}
	if runtime.GOOS == "darwin" && !strings.Contains(gitAction.Command, "brew install git") {
		t.Errorf("expected brew install git on darwin, got %s", gitAction.Command)
	}
}

func TestPlan_UnknownTool(t *testing.T) {
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
		return "", fmt.Errorf("not found")
	}
	runCommand = func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("not found")
	}
	getenv = func(key string) string { return "" }
	userHomeDir = func() (string, error) { return "/tmp/nonexistent-home-for-test", nil }

	saved := &Snapshot{
		Tools: []ToolInfo{
			{Name: "some-exotic-tool", Version: "1.0.0"},
		},
	}

	result := Plan(saved)

	found := false
	for _, a := range result.Actions {
		if a.Tool == "some-exotic-tool" {
			found = true
			if a.Status != "skipped" {
				t.Errorf("expected skipped for unknown tool, got %s", a.Status)
			}
			if !strings.Contains(a.Message, "no known install command") {
				t.Errorf("expected 'no known install command' message, got %s", a.Message)
			}
		}
	}
	if !found {
		t.Error("expected action for some-exotic-tool")
	}
}

func TestPlan_NoMissingTools(t *testing.T) {
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
		return "", fmt.Errorf("not found")
	}
	runCommand = func(name string, args ...string) ([]byte, error) {
		if name == "git" {
			return []byte("git version 2.42.0"), nil
		}
		return nil, fmt.Errorf("not found")
	}
	getenv = func(key string) string { return "" }
	userHomeDir = func() (string, error) { return "/tmp/nonexistent-home-for-test", nil }

	saved := &Snapshot{
		Tools: []ToolInfo{
			{Name: "git", Version: "git version 2.42.0"},
		},
	}

	result := Plan(saved)

	// git matches, so no actions for tools.
	for _, a := range result.Actions {
		if a.Tool == "git" {
			t.Error("did not expect action for git when it already exists")
		}
	}
}

func TestPlan_MissingBrewPackages(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping on non-darwin")
	}

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
		if file == "brew" {
			return "/opt/homebrew/bin/brew", nil
		}
		return "", fmt.Errorf("not found")
	}
	runCommand = func(name string, args ...string) ([]byte, error) {
		if name == "brew" && len(args) > 0 && args[0] == "list" {
			return []byte("git\nwget\n"), nil
		}
		return nil, fmt.Errorf("not found")
	}
	getenv = func(key string) string { return "" }
	userHomeDir = func() (string, error) { return "/tmp/nonexistent-home-for-test", nil }

	saved := &Snapshot{
		Brew: []string{"git", "wget", "jq", "htop"},
	}

	result := Plan(saved)

	brewActions := 0
	for _, a := range result.Actions {
		if a.Tool == "jq" || a.Tool == "htop" {
			brewActions++
			if !strings.Contains(a.Command, "brew install") {
				t.Errorf("expected brew install for %s, got %s", a.Tool, a.Command)
			}
		}
	}
	if brewActions != 2 {
		t.Errorf("expected 2 brew install actions, got %d", brewActions)
	}
}

func TestExecute_DryRun(t *testing.T) {
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
		return "", fmt.Errorf("not found")
	}
	runCommand = func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("not found")
	}
	getenv = func(key string) string { return "" }
	userHomeDir = func() (string, error) { return "/tmp/nonexistent-home-for-test", nil }

	saved := &Snapshot{
		Tools: []ToolInfo{
			{Name: "git", Version: "git version 2.42.0"},
		},
	}

	result := Execute(saved, true)

	// Dry run should not change status from pending.
	for _, a := range result.Actions {
		if a.Tool == "git" && a.Status != "pending" && a.Status != "skipped" {
			t.Errorf("expected pending or skipped in dry run, got %s", a.Status)
		}
	}
}

func TestExecute_InstallSuccess(t *testing.T) {
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

	// For Capture(): no tools found.
	lookPath = func(file string) (string, error) {
		return "", fmt.Errorf("not found")
	}
	getenv = func(key string) string { return "" }
	userHomeDir = func() (string, error) { return "/tmp/nonexistent-home-for-test", nil }

	installCalled := false
	runCommand = func(name string, args ...string) ([]byte, error) {
		// The install command execution.
		if name == "brew" || name == "sudo" {
			installCalled = true
			return []byte("installed"), nil
		}
		return nil, fmt.Errorf("not found")
	}

	saved := &Snapshot{
		Tools: []ToolInfo{
			{Name: "git", Version: "git version 2.42.0"},
		},
	}

	result := Execute(saved, false)

	if !installCalled {
		// On non-darwin, the install command might differ.
		if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
			t.Error("expected install command to be called")
		}
	}

	if result.Summary.Installed+result.Summary.Skipped+result.Summary.Failed == 0 {
		t.Error("expected at least one action result")
	}
}

func TestExecute_InstallFailure(t *testing.T) {
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
		return "", fmt.Errorf("not found")
	}
	getenv = func(key string) string { return "" }
	userHomeDir = func() (string, error) { return "/tmp/nonexistent-home-for-test", nil }

	runCommand = func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("install failed: permission denied")
	}

	saved := &Snapshot{
		Tools: []ToolInfo{
			{Name: "git", Version: "git version 2.42.0"},
		},
	}

	result := Execute(saved, false)

	hasFailed := false
	for _, a := range result.Actions {
		if a.Status == "failed" {
			hasFailed = true
		}
	}

	// On supported OS, should have a failed action.
	if (runtime.GOOS == "darwin" || runtime.GOOS == "linux") && !hasFailed {
		t.Error("expected at least one failed action")
	}
}

func TestFormatPlan_Empty(t *testing.T) {
	result := &SetupResult{}
	output := FormatPlan(result)

	if !strings.Contains(output, "Nothing to install") {
		t.Error("expected 'Nothing to install' for empty plan")
	}
}

func TestFormatPlan_WithActions(t *testing.T) {
	result := &SetupResult{
		Actions: []SetupAction{
			{Tool: "git", Command: "brew install git", Status: "pending"},
			{Tool: "exotic", Status: "skipped", Message: "no known install command"},
			{Tool: "curl", Status: "installed"},
			{Tool: "make", Status: "failed", Message: "permission denied"},
		},
	}

	output := FormatPlan(result)

	if !strings.Contains(output, "brew install git") {
		t.Error("expected pending action with command")
	}
	if !strings.Contains(output, "skipped") {
		t.Error("expected skipped action")
	}
	if !strings.Contains(output, "installed") {
		t.Error("expected installed action")
	}
	if !strings.Contains(output, "FAILED") {
		t.Error("expected failed action")
	}
}
