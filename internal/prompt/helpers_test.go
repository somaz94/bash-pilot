package prompt

import (
	"fmt"
	"os"
	"testing"
)

func TestGetCurrentUser(t *testing.T) {
	origUser := currentUser
	origHost := hostname
	defer func() {
		currentUser = origUser
		hostname = origHost
	}()

	currentUser = func() string { return "alice" }
	hostname = func() string { return "server1" }

	got := getCurrentUser()
	if got != "alice@server1" {
		t.Errorf("expected alice@server1, got %s", got)
	}
}

func TestGetCurrentDir_Home(t *testing.T) {
	origGetwd := getwd
	defer func() { getwd = origGetwd }()

	home, _ := os.UserHomeDir()
	getwd = func() (string, error) { return home + "/projects/foo", nil }

	got := getCurrentDir()
	if got != "~/projects/foo" {
		t.Errorf("expected ~/projects/foo, got %s", got)
	}
}

func TestGetCurrentDir_NonHome(t *testing.T) {
	origGetwd := getwd
	defer func() { getwd = origGetwd }()

	getwd = func() (string, error) { return "/tmp/something", nil }

	got := getCurrentDir()
	if got != "/tmp/something" {
		t.Errorf("expected /tmp/something, got %s", got)
	}
}

func TestGetCurrentDir_Error(t *testing.T) {
	origGetwd := getwd
	defer func() { getwd = origGetwd }()

	getwd = func() (string, error) { return "", fmt.Errorf("error") }

	got := getCurrentDir()
	if got != "~" {
		t.Errorf("expected ~ on error, got %s", got)
	}
}

func TestGetGitBranch_Normal(t *testing.T) {
	origRun := runCommand
	defer func() { runCommand = origRun }()

	runCommand = func(name string, args ...string) ([]byte, error) {
		if args[0] == "symbolic-ref" {
			return []byte("feature/test\n"), nil
		}
		return nil, fmt.Errorf("unexpected")
	}

	got := getGitBranch()
	if got != "feature/test" {
		t.Errorf("expected feature/test, got %s", got)
	}
}

func TestGetGitBranch_DetachedTag(t *testing.T) {
	origRun := runCommand
	defer func() { runCommand = origRun }()

	runCommand = func(name string, args ...string) ([]byte, error) {
		if args[0] == "symbolic-ref" {
			return nil, fmt.Errorf("not on branch")
		}
		if args[0] == "describe" {
			return []byte("v1.2.3\n"), nil
		}
		return nil, fmt.Errorf("unexpected")
	}

	got := getGitBranch()
	if got != "v1.2.3" {
		t.Errorf("expected v1.2.3, got %s", got)
	}
}

func TestGetGitBranch_NoRepo(t *testing.T) {
	origRun := runCommand
	defer func() { runCommand = origRun }()

	runCommand = func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("not a git repo")
	}

	got := getGitBranch()
	if got != "" {
		t.Errorf("expected empty string, got %s", got)
	}
}

func TestIsGitDirty_Clean(t *testing.T) {
	origRun := runCommand
	defer func() { runCommand = origRun }()

	runCommand = func(name string, args ...string) ([]byte, error) {
		return nil, nil // git diff --quiet succeeds = clean
	}

	if isGitDirty() {
		t.Error("expected clean repo")
	}
}

func TestIsGitDirty_UnstagedChanges(t *testing.T) {
	origRun := runCommand
	defer func() { runCommand = origRun }()

	runCommand = func(name string, args ...string) ([]byte, error) {
		if len(args) >= 2 && args[1] == "--quiet" {
			return nil, fmt.Errorf("exit 1") // unstaged changes
		}
		return nil, nil
	}

	if !isGitDirty() {
		t.Error("expected dirty repo")
	}
}

func TestIsGitDirty_StagedChanges(t *testing.T) {
	origRun := runCommand
	defer func() { runCommand = origRun }()

	callCount := 0
	runCommand = func(name string, args ...string) ([]byte, error) {
		callCount++
		if callCount == 1 {
			return nil, nil // git diff --quiet ok
		}
		return nil, fmt.Errorf("exit 1") // git diff --cached --quiet fails
	}

	if !isGitDirty() {
		t.Error("expected dirty with staged changes")
	}
}

func TestGetK8sContext_NoKubectl(t *testing.T) {
	origLookPath := lookPath
	defer func() { lookPath = origLookPath }()

	lookPath = func(file string) (string, error) {
		return "", fmt.Errorf("not found")
	}

	got := getK8sContext()
	if got != "" {
		t.Errorf("expected empty without kubectl, got %s", got)
	}
}

func TestGetK8sContext_WithContext(t *testing.T) {
	origLookPath := lookPath
	origRun := runCommand
	defer func() {
		lookPath = origLookPath
		runCommand = origRun
	}()

	lookPath = func(file string) (string, error) { return "/usr/bin/kubectl", nil }
	runCommand = func(name string, args ...string) ([]byte, error) {
		if name == "kubectl" && len(args) >= 2 {
			if args[1] == "current-context" {
				return []byte("prod-cluster\n"), nil
			}
			if args[1] == "view" {
				return []byte("default"), nil
			}
		}
		return nil, fmt.Errorf("unexpected")
	}

	got := getK8sContext()
	if got != "prod-cluster" {
		t.Errorf("expected prod-cluster, got %s", got)
	}
}

func TestGetK8sContext_WithNamespace(t *testing.T) {
	origLookPath := lookPath
	origRun := runCommand
	defer func() {
		lookPath = origLookPath
		runCommand = origRun
	}()

	lookPath = func(file string) (string, error) { return "/usr/bin/kubectl", nil }
	runCommand = func(name string, args ...string) ([]byte, error) {
		if name == "kubectl" && len(args) >= 2 {
			if args[1] == "current-context" {
				return []byte("staging\n"), nil
			}
			if args[1] == "view" {
				return []byte("monitoring"), nil
			}
		}
		return nil, fmt.Errorf("unexpected")
	}

	got := getK8sContext()
	if got != "staging:monitoring" {
		t.Errorf("expected staging:monitoring, got %s", got)
	}
}

func TestGetK8sContext_NoContext(t *testing.T) {
	origLookPath := lookPath
	origRun := runCommand
	defer func() {
		lookPath = origLookPath
		runCommand = origRun
	}()

	lookPath = func(file string) (string, error) { return "/usr/bin/kubectl", nil }
	runCommand = func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("no context")
	}

	got := getK8sContext()
	if got != "" {
		t.Errorf("expected empty, got %s", got)
	}
}

func TestGetK8sContext_NamespaceError(t *testing.T) {
	origLookPath := lookPath
	origRun := runCommand
	defer func() {
		lookPath = origLookPath
		runCommand = origRun
	}()

	lookPath = func(file string) (string, error) { return "/usr/bin/kubectl", nil }
	runCommand = func(name string, args ...string) ([]byte, error) {
		if name == "kubectl" && len(args) >= 2 {
			if args[1] == "current-context" {
				return []byte("my-cluster\n"), nil
			}
			if args[1] == "view" {
				return nil, fmt.Errorf("error")
			}
		}
		return nil, fmt.Errorf("unexpected")
	}

	got := getK8sContext()
	if got != "my-cluster" {
		t.Errorf("expected my-cluster (no namespace), got %s", got)
	}
}
