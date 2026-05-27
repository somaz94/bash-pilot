package report

import "strings"

// Severity values used across modules (ssh.Audit, git.Doctor, env.Check).
// "fail" is an alias for "error" kept for backward compatibility with ssh.AuditSeverity,
// which predates this unified set.
const (
	SeverityOK    = "ok"
	SeverityWarn  = "warn"
	SeverityError = "error"
	SeverityFail  = "fail"
)

// RenderSeverity returns the formatted line for a severity string.
// Accepts "ok" / "warn" / "error" / "fail" (case-insensitive); unknown
// values fall back to the plain text so callers do not silently drop output.
func (f *Formatter) RenderSeverity(severity, text string) string {
	switch strings.ToLower(severity) {
	case SeverityOK:
		return f.OK(text)
	case SeverityWarn:
		return f.Warn(text)
	case SeverityError, SeverityFail:
		return f.Fail(text)
	default:
		return text
	}
}
