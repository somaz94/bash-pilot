package prompt

import (
	"fmt"
	"strings"
	"testing"
)

func TestGenerateScript_Minimal(t *testing.T) {
	script := GenerateScript(Options{Theme: ThemeMinimal})

	if !strings.Contains(script, "__bp_git_branch") {
		t.Error("expected git branch function in script")
	}
	if strings.Contains(script, "__bp_k8s_context") {
		t.Error("minimal theme should not include k8s context")
	}
	if !strings.Contains(script, "PROMPT_COMMAND=__bp_prompt") {
		t.Error("expected PROMPT_COMMAND assignment")
	}
	if !strings.Contains(script, "bash-pilot prompt init") {
		t.Error("expected usage comment in header")
	}
}

func TestGenerateScript_Full(t *testing.T) {
	script := GenerateScript(Options{Theme: ThemeFull})

	if !strings.Contains(script, "__bp_git_branch") {
		t.Error("expected git branch function")
	}
	if !strings.Contains(script, "__bp_k8s_context") {
		t.Error("full theme should include k8s context")
	}
	if !strings.Contains(script, "k8s_info") {
		t.Error("expected k8s_info in PS1 builder")
	}
}

func TestGenerateScript_FullNoK8s(t *testing.T) {
	script := GenerateScript(Options{Theme: ThemeFull, NoK8s: true})

	if strings.Contains(script, "__bp_k8s_context") {
		t.Error("--no-k8s should exclude k8s context even in full theme")
	}
}

func TestGenerateScript_DefaultMinimal(t *testing.T) {
	script := GenerateScript(Options{})

	// Default theme is empty string, which is not ThemeFull
	if strings.Contains(script, "__bp_k8s_context") {
		t.Error("default should be minimal (no k8s)")
	}
}

func TestShowComponents_WithGit(t *testing.T) {
	origRun := runCommand
	origGetwd := getwd
	origUser := currentUser
	origHost := hostname
	defer func() {
		runCommand = origRun
		getwd = origGetwd
		currentUser = origUser
		hostname = origHost
	}()

	currentUser = func() string { return "testuser" }
	hostname = func() string { return "testhost" }
	getwd = func() (string, error) { return "/home/testuser/project", nil }

	callCount := 0
	runCommand = func(name string, args ...string) ([]byte, error) {
		if name == "git" && len(args) > 0 {
			switch args[0] {
			case "symbolic-ref":
				return []byte("main\n"), nil
			case "diff":
				callCount++
				if callCount == 1 {
					// git diff --quiet → dirty
					return nil, fmt.Errorf("exit 1")
				}
				return nil, nil
			}
		}
		return nil, fmt.Errorf("not found")
	}

	components := ShowComponents(Options{Theme: ThemeMinimal})

	if len(components) < 3 {
		t.Fatalf("expected at least 3 components, got %d", len(components))
	}

	// user@host
	if components[0].Value != "testuser@testhost" {
		t.Errorf("expected testuser@testhost, got %s", components[0].Value)
	}

	// directory
	if components[1].Name != "directory" {
		t.Errorf("expected directory component, got %s", components[1].Name)
	}

	// git
	if components[2].Name != "git" {
		t.Errorf("expected git component, got %s", components[2].Name)
	}
	if !strings.Contains(components[2].Value, "main") {
		t.Errorf("expected main branch, got %s", components[2].Value)
	}
	if !strings.Contains(components[2].Value, "*") {
		t.Errorf("expected dirty indicator, got %s", components[2].Value)
	}
}

func TestShowComponents_NoGit(t *testing.T) {
	origRun := runCommand
	origGetwd := getwd
	origUser := currentUser
	origHost := hostname
	defer func() {
		runCommand = origRun
		getwd = origGetwd
		currentUser = origUser
		hostname = origHost
	}()

	currentUser = func() string { return "user" }
	hostname = func() string { return "host" }
	getwd = func() (string, error) { return "/tmp", nil }
	runCommand = func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("not a git repo")
	}

	components := ShowComponents(Options{Theme: ThemeMinimal})

	// Should have user@host and directory only
	if len(components) != 2 {
		t.Errorf("expected 2 components (no git), got %d", len(components))
	}
}

func TestShowComponents_FullWithK8s(t *testing.T) {
	origRun := runCommand
	origLookPath := lookPath
	origGetwd := getwd
	origUser := currentUser
	origHost := hostname
	defer func() {
		runCommand = origRun
		lookPath = origLookPath
		getwd = origGetwd
		currentUser = origUser
		hostname = origHost
	}()

	currentUser = func() string { return "user" }
	hostname = func() string { return "host" }
	getwd = func() (string, error) { return "/tmp", nil }
	lookPath = func(file string) (string, error) {
		if file == "kubectl" {
			return "/usr/bin/kubectl", nil
		}
		return "", fmt.Errorf("not found")
	}
	runCommand = func(name string, args ...string) ([]byte, error) {
		if name == "git" {
			return nil, fmt.Errorf("not a git repo")
		}
		if name == "kubectl" && len(args) >= 2 {
			if args[1] == "current-context" {
				return []byte("my-cluster\n"), nil
			}
			if args[1] == "view" {
				return []byte("kube-system"), nil
			}
		}
		return nil, fmt.Errorf("unexpected")
	}

	components := ShowComponents(Options{Theme: ThemeFull})

	found := false
	for _, c := range components {
		if c.Name == "k8s" {
			found = true
			if c.Value != "my-cluster:kube-system" {
				t.Errorf("expected my-cluster:kube-system, got %s", c.Value)
			}
		}
	}
	if !found {
		t.Error("expected k8s component in full theme")
	}
}

func TestShowComponents_FullNoK8sFlag(t *testing.T) {
	origRun := runCommand
	origGetwd := getwd
	origUser := currentUser
	origHost := hostname
	defer func() {
		runCommand = origRun
		getwd = origGetwd
		currentUser = origUser
		hostname = origHost
	}()

	currentUser = func() string { return "user" }
	hostname = func() string { return "host" }
	getwd = func() (string, error) { return "/tmp", nil }
	runCommand = func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("not found")
	}

	components := ShowComponents(Options{Theme: ThemeFull, NoK8s: true})

	for _, c := range components {
		if c.Name == "k8s" {
			t.Error("expected no k8s component with --no-k8s")
		}
	}
}

func TestFormatPreview(t *testing.T) {
	components := []Component{
		{Name: "user@host", Value: "user@host"},
		{Name: "directory", Value: "~/project"},
		{Name: "git", Value: "main *"},
	}

	result := FormatPreview(components)

	if !strings.Contains(result, "user@host") {
		t.Error("expected user@host in preview")
	}
	if !strings.Contains(result, "~/project") {
		t.Error("expected directory in preview")
	}
	if !strings.Contains(result, "main *") {
		t.Error("expected git info in preview")
	}
}

func TestFormatPreview_Empty(t *testing.T) {
	result := FormatPreview(nil)
	if result != "" {
		t.Errorf("expected empty string for nil components, got %q", result)
	}
}
