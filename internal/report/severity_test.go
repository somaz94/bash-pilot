package report

import (
	"bytes"
	"strings"
	"testing"
)

func TestRenderSeverity(t *testing.T) {
	// Use plain format so the asserts don't have to deal with ANSI escapes.
	f := NewFormatter(&bytes.Buffer{}, "plain", false)

	tests := []struct {
		name     string
		severity string
		text     string
		wantMark string
	}{
		{"ok", SeverityOK, "all good", "✓"},
		{"warn", SeverityWarn, "heads up", "!"},
		{"error", SeverityError, "broken", "✗"},
		{"fail (ssh alias)", SeverityFail, "broken", "✗"},
		{"uppercase OK", "OK", "all good", "✓"},
		{"mixed Warn", "Warn", "heads up", "!"},
		{"unknown falls back to plain", "moot", "msg", "msg"},
		{"empty falls back to plain", "", "msg", "msg"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := f.RenderSeverity(tc.severity, tc.text)
			if !strings.Contains(got, tc.wantMark) {
				t.Errorf("RenderSeverity(%q, %q) = %q, want substring %q",
					tc.severity, tc.text, got, tc.wantMark)
			}
			if !strings.Contains(got, tc.text) && tc.severity != "" {
				t.Errorf("RenderSeverity(%q, %q) = %q, missing message", tc.severity, tc.text, got)
			}
		})
	}
}
