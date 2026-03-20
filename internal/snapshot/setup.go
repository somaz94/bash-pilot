package snapshot

import (
	"fmt"
	"runtime"
	"strings"
)

// SetupAction represents a single install action.
type SetupAction struct {
	Tool    string `json:"tool"`
	Command string `json:"command"`
	Status  string `json:"status"` // "pending", "installed", "skipped", "failed"
	Message string `json:"message,omitempty"`
}

// SetupResult holds the result of a setup operation.
type SetupResult struct {
	Actions []SetupAction `json:"actions"`
	Summary SetupSummary  `json:"summary"`
}

// SetupSummary holds counts.
type SetupSummary struct {
	Installed int `json:"installed"`
	Skipped   int `json:"skipped"`
	Failed    int `json:"failed"`
}

// installCommands maps tool names to install commands per OS.
var installCommands = map[string]map[string][]string{
	"git": {
		"darwin": {"brew", "install", "git"},
		"linux":  {"sudo", "apt-get", "install", "-y", "git"},
	},
	"curl": {
		"darwin": {"brew", "install", "curl"},
		"linux":  {"sudo", "apt-get", "install", "-y", "curl"},
	},
	"make": {
		"darwin": {"brew", "install", "make"},
		"linux":  {"sudo", "apt-get", "install", "-y", "make"},
	},
	"docker": {
		"darwin": {"brew", "install", "--cask", "docker"},
		"linux":  {"sudo", "apt-get", "install", "-y", "docker.io"},
	},
	"kubectl": {
		"darwin": {"brew", "install", "kubectl"},
		"linux":  {"sudo", "snap", "install", "kubectl", "--classic"},
	},
	"helm": {
		"darwin": {"brew", "install", "helm"},
		"linux":  {"sudo", "snap", "install", "helm", "--classic"},
	},
	"terraform": {
		"darwin": {"brew", "install", "terraform"},
		"linux":  {"sudo", "snap", "install", "terraform", "--classic"},
	},
	"go": {
		"darwin": {"brew", "install", "go"},
		"linux":  {"sudo", "snap", "install", "go", "--classic"},
	},
	"node": {
		"darwin": {"brew", "install", "node"},
		"linux":  {"sudo", "apt-get", "install", "-y", "nodejs"},
	},
	"python3": {
		"darwin": {"brew", "install", "python3"},
		"linux":  {"sudo", "apt-get", "install", "-y", "python3"},
	},
	"ssh": {
		"darwin": {"brew", "install", "openssh"},
		"linux":  {"sudo", "apt-get", "install", "-y", "openssh-client"},
	},
}

// Plan returns the list of actions needed to match the snapshot, without executing.
func Plan(saved *Snapshot) *SetupResult {
	diff := Diff(saved)
	result := &SetupResult{}
	osName := runtime.GOOS

	// Find missing tools from diff.
	for _, sec := range diff.Sections {
		if sec.Name != "Tools" {
			continue
		}
		for _, e := range sec.Entries {
			if e.Status != "missing" {
				continue
			}
			action := SetupAction{
				Tool:   e.Key,
				Status: "pending",
			}
			cmds, ok := installCommands[e.Key]
			if !ok {
				action.Status = "skipped"
				action.Message = "no known install command"
				action.Command = ""
				result.Summary.Skipped++
			} else if cmdArgs, osOk := cmds[osName]; osOk {
				action.Command = strings.Join(cmdArgs, " ")
			} else {
				action.Status = "skipped"
				action.Message = fmt.Sprintf("no install command for %s", osName)
				action.Command = ""
				result.Summary.Skipped++
			}
			result.Actions = append(result.Actions, action)
		}
	}

	// Find missing brew packages.
	for _, sec := range diff.Sections {
		if sec.Name != "Brew Packages" {
			continue
		}
		if osName != "darwin" {
			continue
		}
		for _, e := range sec.Entries {
			if e.Status != "missing" {
				continue
			}
			result.Actions = append(result.Actions, SetupAction{
				Tool:    e.Key,
				Command: fmt.Sprintf("brew install %s", e.Key),
				Status:  "pending",
			})
		}
	}

	return result
}

// Execute runs the planned setup actions, installing missing tools.
func Execute(saved *Snapshot, dryRun bool) *SetupResult {
	plan := Plan(saved)

	if dryRun {
		return plan
	}

	for i := range plan.Actions {
		action := &plan.Actions[i]
		if action.Status != "pending" {
			continue
		}

		args := strings.Fields(action.Command)
		if len(args) == 0 {
			action.Status = "skipped"
			action.Message = "empty command"
			plan.Summary.Skipped++
			continue
		}

		_, err := runCommand(args[0], args[1:]...)
		if err != nil {
			action.Status = "failed"
			action.Message = err.Error()
			plan.Summary.Failed++
		} else {
			action.Status = "installed"
			action.Message = "success"
			plan.Summary.Installed++
		}
	}

	return plan
}

// FormatPlan returns a human-readable plan summary.
func FormatPlan(result *SetupResult) string {
	var b strings.Builder

	if len(result.Actions) == 0 {
		b.WriteString("  Nothing to install — environment matches snapshot.\n")
		return b.String()
	}

	for _, a := range result.Actions {
		switch a.Status {
		case "pending":
			b.WriteString(fmt.Sprintf("  %-20s %s\n", a.Tool, a.Command))
		case "installed":
			b.WriteString(fmt.Sprintf("  %-20s installed\n", a.Tool))
		case "skipped":
			b.WriteString(fmt.Sprintf("  %-20s skipped (%s)\n", a.Tool, a.Message))
		case "failed":
			b.WriteString(fmt.Sprintf("  %-20s FAILED (%s)\n", a.Tool, a.Message))
		}
	}

	return b.String()
}
