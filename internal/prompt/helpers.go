package prompt

import (
	"os"
	"os/exec"
	"os/user"
	"strings"
)

// Function variables for testing.
var (
	runCommand = func(name string, args ...string) ([]byte, error) {
		return exec.Command(name, args...).Output()
	}
	lookPath   = exec.LookPath
	getwd      = os.Getwd
	currentUser = func() string {
		u, err := user.Current()
		if err != nil {
			return "user"
		}
		return u.Username
	}
	hostname = func() string {
		h, err := os.Hostname()
		if err != nil {
			return "host"
		}
		return h
	}
)

func getCurrentUser() string {
	return currentUser() + "@" + hostname()
}

func getCurrentDir() string {
	dir, err := getwd()
	if err != nil {
		return "~"
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return dir
	}
	if strings.HasPrefix(dir, home) {
		return "~" + dir[len(home):]
	}
	return dir
}

func getGitBranch() string {
	out, err := runCommand("git", "symbolic-ref", "--short", "HEAD")
	if err != nil {
		// Try detached HEAD (tag).
		out, err = runCommand("git", "describe", "--tags", "--exact-match")
		if err != nil {
			return ""
		}
	}
	return strings.TrimSpace(string(out))
}

func isGitDirty() bool {
	_, err := runCommand("git", "diff", "--quiet")
	if err != nil {
		return true
	}
	_, err = runCommand("git", "diff", "--cached", "--quiet")
	return err != nil
}

func getK8sContext() string {
	if _, err := lookPath("kubectl"); err != nil {
		return ""
	}

	out, err := runCommand("kubectl", "config", "current-context")
	if err != nil {
		return ""
	}
	ctx := strings.TrimSpace(string(out))

	out, err = runCommand("kubectl", "config", "view", "--minify", "--output", "jsonpath={..namespace}")
	if err != nil {
		return ctx
	}
	ns := strings.TrimSpace(string(out))

	if ns != "" && ns != "default" {
		return ctx + ":" + ns
	}
	return ctx
}
