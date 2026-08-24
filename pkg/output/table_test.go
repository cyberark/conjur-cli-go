package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestRenderTableCollection(t *testing.T) {
	var buf bytes.Buffer
	data := []map[string]any{
		{"id": "alice", "kind": "user"},
		{"id": "bob", "kind": "user"},
	}
	if err := renderTable(&buf, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3 (header + 2 rows): %q", len(lines), got)
	}
	if !strings.Contains(lines[0], "ID") || !strings.Contains(lines[0], "KIND") {
		t.Fatalf("header missing columns: %q", lines[0])
	}
	if !strings.Contains(lines[1], "alice") || !strings.Contains(lines[2], "bob") {
		t.Fatalf("rows missing values: %q", got)
	}
}

func TestRenderTableSingleObject(t *testing.T) {
	var buf bytes.Buffer
	if err := renderTable(&buf, map[string]any{"id": "alice", "kind": "user"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "alice") {
		t.Fatalf("got %q", got)
	}
}

func TestRenderTableScalarSlice(t *testing.T) {
	var buf bytes.Buffer
	if err := renderTable(&buf, []string{"a", "b"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3: %q", len(lines), got)
	}
	if !strings.Contains(lines[0], "VALUE") {
		t.Fatalf("header missing VALUE column: %q", lines[0])
	}
}

func TestRenderTableWrapsLongValuesWithinTerminalWidth(t *testing.T) {
	orig := terminalSizeFunc
	const width = 60
	terminalSizeFunc = func() (int, int, error) { return width, 24, nil }
	t.Cleanup(func() { terminalSizeFunc = orig })

	var buf bytes.Buffer
	data := map[string]any{
		"payload":   strings.Repeat("A", 80),
		"protected": strings.Repeat("B", 80),
		"signature": strings.Repeat("C", 80),
	}
	if err := renderTable(&buf, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, line := range strings.Split(strings.TrimRight(buf.String(), "\n"), "\n") {
		if len(line) > width {
			t.Fatalf("line exceeds terminal width %d: %q (%d chars)", width, line, len(line))
		}
	}
	if !strings.Contains(buf.String(), "AAAA") || !strings.Contains(buf.String(), "CCCC") {
		t.Fatalf("expected wrapped values to still appear in output: %q", buf.String())
	}
}

func TestRenderTableWrapsWideHeader(t *testing.T) {
	orig := terminalSizeFunc
	terminalSizeFunc = func() (int, int, error) { return 40, 24, nil }
	t.Cleanup(func() { terminalSizeFunc = orig })

	var buf bytes.Buffer
	data := map[string]any{
		"a_very_long_field_name_used_as_a_header": "short",
		"b": "short",
	}
	if err := renderTable(&buf, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, line := range strings.Split(strings.TrimRight(buf.String(), "\n"), "\n") {
		if len(line) > 40 {
			t.Fatalf("line exceeds terminal width 40: %q (%d chars)", line, len(line))
		}
	}
}

func TestRenderTableEmptyCollection(t *testing.T) {
	var buf bytes.Buffer
	if err := renderTable(&buf, []any{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := buf.String(); got != "" {
		t.Fatalf("got %q, want empty output for zero rows", got)
	}
}

func TestParseFormatTable(t *testing.T) {
	f, err := ParseFormat("table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Type != FormatTable {
		t.Fatalf("got %v, want FormatTable", f.Type)
	}
}

func TestRenderDispatchesTable(t *testing.T) {
	var buf bytes.Buffer
	f, _ := ParseFormat("table")
	if err := Render(&buf, f, []map[string]any{{"id": "alice"}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "alice") {
		t.Fatalf("got %q", buf.String())
	}
}
