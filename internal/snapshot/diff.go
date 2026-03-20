package snapshot

import (
	"fmt"
	"strings"
)

// DiffResult holds differences between two snapshots.
type DiffResult struct {
	Sections []DiffSection `json:"sections"`
	Summary  DiffSummary   `json:"summary"`
}

// DiffSection groups diffs by category.
type DiffSection struct {
	Name    string      `json:"name"`
	Entries []DiffEntry `json:"entries"`
}

// DiffEntry represents a single difference.
type DiffEntry struct {
	Status string `json:"status"` // "match", "mismatch", "missing", "extra"
	Key    string `json:"key"`
	Left   string `json:"left,omitempty"`  // snapshot value
	Right  string `json:"right,omitempty"` // current value
}

// DiffSummary holds counts.
type DiffSummary struct {
	Match    int `json:"match"`
	Mismatch int `json:"mismatch"`
	Missing  int `json:"missing"` // in snapshot but not current
	Extra    int `json:"extra"`   // in current but not snapshot
}

// Diff compares a saved snapshot against the current environment.
func Diff(saved *Snapshot) *DiffResult {
	current := Capture()
	result := &DiffResult{}

	// System
	sys := DiffSection{Name: "System"}
	sys.Entries = append(sys.Entries, compareField("OS", saved.OS, current.OS))
	sys.Entries = append(sys.Entries, compareField("Arch", saved.Arch, current.Arch))
	sys.Entries = append(sys.Entries, compareField("Shell", saved.Shell.Shell, current.Shell.Shell))
	sys.Entries = append(sys.Entries, compareField("Editor", saved.Shell.Editor, current.Shell.Editor))
	result.Sections = append(result.Sections, sys)

	// Tools
	tools := DiffSection{Name: "Tools"}
	savedTools := make(map[string]ToolInfo)
	for _, t := range saved.Tools {
		savedTools[t.Name] = t
	}
	currentTools := make(map[string]ToolInfo)
	for _, t := range current.Tools {
		currentTools[t.Name] = t
	}

	// Check all saved tools.
	for _, st := range saved.Tools {
		ct, exists := currentTools[st.Name]
		if !exists {
			tools.Entries = append(tools.Entries, DiffEntry{
				Status: "missing",
				Key:    st.Name,
				Left:   st.Version,
			})
			continue
		}
		if st.Version != ct.Version {
			tools.Entries = append(tools.Entries, DiffEntry{
				Status: "mismatch",
				Key:    st.Name,
				Left:   st.Version,
				Right:  ct.Version,
			})
		} else {
			tools.Entries = append(tools.Entries, DiffEntry{
				Status: "match",
				Key:    st.Name,
				Left:   st.Version,
				Right:  ct.Version,
			})
		}
	}
	// Check for extra tools in current.
	for _, ct := range current.Tools {
		if _, exists := savedTools[ct.Name]; !exists {
			tools.Entries = append(tools.Entries, DiffEntry{
				Status: "extra",
				Key:    ct.Name,
				Right:  ct.Version,
			})
		}
	}
	result.Sections = append(result.Sections, tools)

	// Git
	git := DiffSection{Name: "Git"}
	git.Entries = append(git.Entries, compareField("user.email", saved.Git.Email, current.Git.Email))
	git.Entries = append(git.Entries, compareField("user.name", saved.Git.Name, current.Git.Name))
	result.Sections = append(result.Sections, git)

	// SSH Keys
	ssh := DiffSection{Name: "SSH Keys"}
	savedKeys := make(map[string]SSHKeyInfo)
	for _, k := range saved.SSHKeys {
		savedKeys[k.Name] = k
	}
	currentKeys := make(map[string]SSHKeyInfo)
	for _, k := range current.SSHKeys {
		currentKeys[k.Name] = k
	}
	for _, sk := range saved.SSHKeys {
		ck, exists := currentKeys[sk.Name]
		if !exists {
			ssh.Entries = append(ssh.Entries, DiffEntry{
				Status: "missing",
				Key:    sk.Name,
				Left:   sk.Fingerprint,
			})
		} else if sk.Fingerprint != ck.Fingerprint {
			ssh.Entries = append(ssh.Entries, DiffEntry{
				Status: "mismatch",
				Key:    sk.Name,
				Left:   sk.Fingerprint,
				Right:  ck.Fingerprint,
			})
		} else {
			ssh.Entries = append(ssh.Entries, DiffEntry{
				Status: "match",
				Key:    sk.Name,
			})
		}
	}
	for _, ck := range current.SSHKeys {
		if _, exists := savedKeys[ck.Name]; !exists {
			ssh.Entries = append(ssh.Entries, DiffEntry{
				Status: "extra",
				Key:    ck.Name,
				Right:  ck.Fingerprint,
			})
		}
	}
	result.Sections = append(result.Sections, ssh)

	// K8s Contexts
	if len(saved.K8s) > 0 || len(current.K8s) > 0 {
		k8s := DiffSection{Name: "K8s Contexts"}
		savedCtx := make(map[string]bool)
		for _, c := range saved.K8s {
			savedCtx[c.Name] = true
		}
		currentCtx := make(map[string]bool)
		for _, c := range current.K8s {
			currentCtx[c.Name] = true
		}
		for _, c := range saved.K8s {
			if currentCtx[c.Name] {
				k8s.Entries = append(k8s.Entries, DiffEntry{Status: "match", Key: c.Name})
			} else {
				k8s.Entries = append(k8s.Entries, DiffEntry{Status: "missing", Key: c.Name})
			}
		}
		for _, c := range current.K8s {
			if !savedCtx[c.Name] {
				k8s.Entries = append(k8s.Entries, DiffEntry{Status: "extra", Key: c.Name})
			}
		}
		result.Sections = append(result.Sections, k8s)
	}

	// Brew packages
	if len(saved.Brew) > 0 || len(current.Brew) > 0 {
		brew := DiffSection{Name: "Brew Packages"}
		savedPkgs := make(map[string]bool)
		for _, p := range saved.Brew {
			savedPkgs[p] = true
		}
		currentPkgs := make(map[string]bool)
		for _, p := range current.Brew {
			currentPkgs[p] = true
		}
		for _, p := range saved.Brew {
			if currentPkgs[p] {
				brew.Entries = append(brew.Entries, DiffEntry{Status: "match", Key: p})
			} else {
				brew.Entries = append(brew.Entries, DiffEntry{Status: "missing", Key: p})
			}
		}
		for _, p := range current.Brew {
			if !savedPkgs[p] {
				brew.Entries = append(brew.Entries, DiffEntry{Status: "extra", Key: p})
			}
		}
		result.Sections = append(result.Sections, brew)
	}

	// Compute summary.
	for _, sec := range result.Sections {
		for _, e := range sec.Entries {
			switch e.Status {
			case "match":
				result.Summary.Match++
			case "mismatch":
				result.Summary.Mismatch++
			case "missing":
				result.Summary.Missing++
			case "extra":
				result.Summary.Extra++
			}
		}
	}

	return result
}

func compareField(key, left, right string) DiffEntry {
	if left == right {
		return DiffEntry{Status: "match", Key: key, Left: left, Right: right}
	}
	if left == "" && right != "" {
		return DiffEntry{Status: "extra", Key: key, Right: right}
	}
	if left != "" && right == "" {
		return DiffEntry{Status: "missing", Key: key, Left: left}
	}
	return DiffEntry{Status: "mismatch", Key: key, Left: left, Right: right}
}

// FormatDiff returns a human-readable diff summary.
func FormatDiff(result *DiffResult) string {
	var b strings.Builder

	for _, sec := range result.Sections {
		hasIssues := false
		for _, e := range sec.Entries {
			if e.Status != "match" {
				hasIssues = true
				break
			}
		}
		if !hasIssues {
			b.WriteString(fmt.Sprintf("  %-16s all match\n", sec.Name+":"))
			continue
		}

		b.WriteString(fmt.Sprintf("  %-16s\n", sec.Name+":"))
		for _, e := range sec.Entries {
			switch e.Status {
			case "match":
				// skip matches in detailed view
			case "mismatch":
				b.WriteString(fmt.Sprintf("    ~ %-20s %s → %s\n", e.Key, e.Left, e.Right))
			case "missing":
				b.WriteString(fmt.Sprintf("    - %-20s %s (not in current)\n", e.Key, e.Left))
			case "extra":
				b.WriteString(fmt.Sprintf("    + %-20s %s (new)\n", e.Key, e.Right))
			}
		}
	}

	b.WriteString(fmt.Sprintf("\n  Summary: %d match, %d mismatch, %d missing, %d extra\n",
		result.Summary.Match, result.Summary.Mismatch, result.Summary.Missing, result.Summary.Extra))

	return b.String()
}
