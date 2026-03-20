package report

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewFormatter(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, "color", false)

	if f.Format != "color" {
		t.Errorf("Format = %q, want %q", f.Format, "color")
	}
	if f.NoColor {
		t.Error("NoColor should be false")
	}
}

func TestColor(t *testing.T) {
	var buf bytes.Buffer

	// With color enabled.
	f := NewFormatter(&buf, "color", false)
	got := f.Color(Green, "test")
	if !strings.Contains(got, "test") {
		t.Errorf("Color() = %q, should contain 'test'", got)
	}
	if !strings.Contains(got, "\033[") {
		t.Error("Color() should contain ANSI codes when color enabled")
	}

	// With color disabled.
	f2 := NewFormatter(&buf, "plain", false)
	got2 := f2.Color(Green, "test")
	if got2 != "test" {
		t.Errorf("Color() with plain = %q, want %q", got2, "test")
	}

	// With NoColor flag.
	f3 := NewFormatter(&buf, "color", true)
	got3 := f3.Color(Green, "test")
	if got3 != "test" {
		t.Errorf("Color() with NoColor = %q, want %q", got3, "test")
	}

	// JSON format should strip color.
	f4 := NewFormatter(&buf, "json", false)
	got4 := f4.Color(Green, "test")
	if got4 != "test" {
		t.Errorf("Color() with json = %q, want %q", got4, "test")
	}
}

func TestOKWarnFail(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, "plain", false)

	ok := f.OK("all good")
	if !strings.Contains(ok, "✓") || !strings.Contains(ok, "all good") {
		t.Errorf("OK() = %q, expected check mark + text", ok)
	}

	warn := f.Warn("caution")
	if !strings.Contains(warn, "!") || !strings.Contains(warn, "caution") {
		t.Errorf("Warn() = %q, expected exclamation + text", warn)
	}

	fail := f.Fail("error")
	if !strings.Contains(fail, "✗") || !strings.Contains(fail, "error") {
		t.Errorf("Fail() = %q, expected cross + text", fail)
	}
}

func TestHeaderFooterRow(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, "plain", false)

	f.Header("Test Section")
	f.Row("line 1")
	f.Row("line 2")
	f.Footer()

	output := buf.String()
	if !strings.Contains(output, "Test Section") {
		t.Error("Header should contain section name")
	}
	if !strings.Contains(output, "line 1") {
		t.Error("Row should contain text")
	}
	if strings.Count(output, "\n") < 3 {
		t.Error("expected at least 3 lines (header + 2 rows + footer)")
	}
}

func TestHeaderFooterRow_JSON(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, "json", false)

	// JSON format should suppress Header/Footer/Row.
	f.Header("Test")
	f.Row("data")
	f.Footer()

	if buf.Len() != 0 {
		t.Errorf("JSON format should suppress Header/Footer/Row, got %q", buf.String())
	}
}

func TestPrintln(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, "plain", false)

	f.Println("hello")
	if buf.String() != "hello\n" {
		t.Errorf("Println() = %q, want %q", buf.String(), "hello\n")
	}
}

func TestJSON(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, "json", false)

	data := map[string]string{"key": "value"}
	if err := f.JSON(data); err != nil {
		t.Fatalf("JSON() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"key"`) || !strings.Contains(output, `"value"`) {
		t.Errorf("JSON() = %q, expected key/value", output)
	}
}

func TestTable(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, "plain", false)

	headers := []string{"NAME", "VALUE"}
	rows := [][]string{
		{"foo", "bar"},
		{"longname", "x"},
	}
	f.Table(headers, rows)

	output := buf.String()
	if !strings.Contains(output, "NAME") {
		t.Error("Table should contain header")
	}
	if !strings.Contains(output, "foo") {
		t.Error("Table should contain row data")
	}
	if !strings.Contains(output, "longname") {
		t.Error("Table should contain all rows")
	}
	// Check alignment — separator line should exist.
	if !strings.Contains(output, "──") {
		t.Error("Table should contain separator")
	}
}
