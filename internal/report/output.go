package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// ANSI color codes.
const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
	Gray   = "\033[90m"
)

// Formatter handles output rendering.
type Formatter struct {
	Writer  io.Writer
	Format  string // color, plain, json, table
	NoColor bool
}

// NewFormatter creates a formatter with the given settings.
func NewFormatter(w io.Writer, format string, noColor bool) *Formatter {
	return &Formatter{
		Writer:  w,
		Format:  format,
		NoColor: noColor,
	}
}

// Color wraps text with ANSI color if color is enabled.
func (f *Formatter) Color(color, text string) string {
	if f.NoColor || f.Format == "plain" || f.Format == "json" {
		return text
	}
	return color + text + Reset
}

// OK returns a green check mark with text.
func (f *Formatter) OK(text string) string {
	return f.Color(Green, "✓") + " " + text
}

// Warn returns a yellow exclamation with text.
func (f *Formatter) Warn(text string) string {
	return f.Color(Yellow, "!") + " " + text
}

// Fail returns a red cross with text.
func (f *Formatter) Fail(text string) string {
	return f.Color(Red, "✗") + " " + text
}

// Header prints a section header.
func (f *Formatter) Header(text string) {
	if f.Format == "json" {
		return
	}
	remaining := 50 - len(text)
	if remaining < 3 {
		remaining = 3
	}
	line := strings.Repeat("─", remaining)
	fmt.Fprintf(f.Writer, "%s %s %s\n", f.Color(Cyan, "┌─"), f.Color(Bold, text), f.Color(Cyan, line))
}

// Footer prints a section footer.
func (f *Formatter) Footer() {
	if f.Format == "json" {
		return
	}
	fmt.Fprintf(f.Writer, "%s\n", f.Color(Cyan, "└"+strings.Repeat("─", 55)))
}

// Row prints a line with leading box character.
func (f *Formatter) Row(text string) {
	if f.Format == "json" {
		return
	}
	fmt.Fprintf(f.Writer, "%s %s\n", f.Color(Cyan, "│"), text)
}

// Println prints a line.
func (f *Formatter) Println(text string) {
	fmt.Fprintln(f.Writer, text)
}

// JSON outputs data as JSON.
func (f *Formatter) JSON(v interface{}) error {
	enc := json.NewEncoder(f.Writer)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// Table prints tabular data with aligned columns.
func (f *Formatter) Table(headers []string, rows [][]string) {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print headers.
	for i, h := range headers {
		fmt.Fprintf(f.Writer, "%-*s  ", widths[i], f.Color(Bold, h))
	}
	fmt.Fprintln(f.Writer)

	// Print separator.
	for i := range headers {
		fmt.Fprintf(f.Writer, "%s  ", strings.Repeat("─", widths[i]))
	}
	fmt.Fprintln(f.Writer)

	// Print rows.
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				fmt.Fprintf(f.Writer, "%-*s  ", widths[i], cell)
			}
		}
		fmt.Fprintln(f.Writer)
	}
}
