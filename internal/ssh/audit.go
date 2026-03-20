package ssh

import (
	"fmt"
	"os"
	"path/filepath"
)

// AuditSeverity represents the severity level of an audit finding.
type AuditSeverity string

const (
	SeverityOK   AuditSeverity = "ok"
	SeverityWarn AuditSeverity = "warn"
	SeverityFail AuditSeverity = "fail"
)

// AuditFinding represents a single security audit result.
type AuditFinding struct {
	Severity AuditSeverity `json:"severity"`
	Key      string        `json:"key"`
	Message  string        `json:"message"`
}

// AuditResult holds all audit findings.
type AuditResult struct {
	Findings []AuditFinding `json:"findings"`
}

// Audit performs a security audit on SSH hosts and their keys.
func Audit(hosts []Host) AuditResult {
	var result AuditResult

	// Check for shared identity files.
	keyUsage := make(map[string][]string)
	for _, h := range hosts {
		if h.IdentityFile != "" {
			keyUsage[h.IdentityFile] = append(keyUsage[h.IdentityFile], h.Name)
		}
	}

	for keyFile, hostNames := range keyUsage {
		keyName := filepath.Base(keyFile)
		if len(hostNames) > 3 {
			result.Findings = append(result.Findings, AuditFinding{
				Severity: SeverityWarn,
				Key:      keyName,
				Message:  fmt.Sprintf("%s: used by %d hosts (consider per-host keys)", keyName, len(hostNames)),
			})
		} else {
			result.Findings = append(result.Findings, AuditFinding{
				Severity: SeverityOK,
				Key:      keyName,
				Message:  fmt.Sprintf("%s: used by %d host(s)", keyName, len(hostNames)),
			})
		}
	}

	// Check file permissions on key files.
	checkedFiles := make(map[string]bool)
	for _, h := range hosts {
		if h.IdentityFile == "" || checkedFiles[h.IdentityFile] {
			continue
		}
		checkedFiles[h.IdentityFile] = true

		info, err := os.Stat(h.IdentityFile)
		if err != nil {
			if os.IsNotExist(err) {
				result.Findings = append(result.Findings, AuditFinding{
					Severity: SeverityFail,
					Key:      filepath.Base(h.IdentityFile),
					Message:  fmt.Sprintf("%s: key file not found", filepath.Base(h.IdentityFile)),
				})
			}
			continue
		}

		perm := info.Mode().Perm()
		if perm&0o077 != 0 {
			result.Findings = append(result.Findings, AuditFinding{
				Severity: SeverityWarn,
				Key:      filepath.Base(h.IdentityFile),
				Message:  fmt.Sprintf("%s: permissions %04o (should be 0600)", filepath.Base(h.IdentityFile), perm),
			})
		} else {
			result.Findings = append(result.Findings, AuditFinding{
				Severity: SeverityOK,
				Key:      filepath.Base(h.IdentityFile),
				Message:  fmt.Sprintf("%s: permissions OK (%04o)", filepath.Base(h.IdentityFile), perm),
			})
		}
	}

	// Check for hosts without identity files.
	for _, h := range hosts {
		if h.IdentityFile == "" {
			result.Findings = append(result.Findings, AuditFinding{
				Severity: SeverityWarn,
				Key:      h.Name,
				Message:  "no IdentityFile specified (will use default keys)",
			})
		}
	}

	return result
}
