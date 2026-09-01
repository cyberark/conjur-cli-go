package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	data := map[string]any{"name": "foo"}

	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		f, _ := ParseFormat("json")
		if err := Render(&buf, f, data); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(buf.String(), "\"name\": \"foo\"") {
			t.Fatalf("got %q", buf.String())
		}
	})

	t.Run("yaml", func(t *testing.T) {
		var buf bytes.Buffer
		f, _ := ParseFormat("yaml")
		if err := Render(&buf, f, data); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if buf.String() != "name: foo\n" {
			t.Fatalf("got %q", buf.String())
		}
	})

	t.Run("custom-columns", func(t *testing.T) {
		var buf bytes.Buffer
		f, err := ParseFormat("custom-columns=NAME:.name")
		if err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if err := Render(&buf, f, data); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(buf.String(), "foo") {
			t.Fatalf("got %q", buf.String())
		}
	})

	t.Run("go-template", func(t *testing.T) {
		var buf bytes.Buffer
		f, err := ParseFormat("go-template={{ .name }}")
		if err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if err := Render(&buf, f, data); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if buf.String() != "foo\n" {
			t.Fatalf("got %q", buf.String())
		}
	})

	t.Run("plain is rejected", func(t *testing.T) {
		var buf bytes.Buffer
		f, _ := ParseFormat("plain")
		if err := Render(&buf, f, data); err == nil {
			t.Fatalf("expected error for FormatPlain, got nil")
		}
	})

	t.Run("go-template with nil ParsedTemplate errors instead of panicking", func(t *testing.T) {
		var buf bytes.Buffer
		// Constructed directly (not via ParseFormat) so ParsedTemplate is nil.
		f := Format{Type: FormatGoTemplate}
		if err := Render(&buf, f, data); err == nil {
			t.Fatalf("expected error for nil ParsedTemplate, got nil")
		}
	})
}

// testItem and testResult are real structs with json tags, used to verify
// that Render works against actual JSON-marshalable structs (not just
// map[string]any), per the package's core design goal.
type testItem struct {
	Name string `json:"name"`
}

type testResult struct {
	ID    string     `json:"id"`
	Count int        `json:"count"`
	Items []testItem `json:"items"`
}

func TestRenderWithRealStruct(t *testing.T) {
	result := testResult{
		ID:    "abc",
		Count: 3,
		Items: []testItem{{Name: "x"}, {Name: "y"}},
	}

	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		f, _ := ParseFormat("json")
		if err := Render(&buf, f, result); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "\"id\": \"abc\"") || !strings.Contains(out, "\"count\": 3") {
			t.Fatalf("got %q", out)
		}
	})

	t.Run("custom-columns", func(t *testing.T) {
		var buf bytes.Buffer
		// Dot-path uses the JSON tag name ".name", not the Go field name
		// ".Name" -- confirming the JSON round-trip model.
		f, err := ParseFormat("custom-columns=ID:.id,FIRST_ITEM:.items[0].name")
		if err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if err := Render(&buf, f, result); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "abc") || !strings.Contains(out, "x") {
			t.Fatalf("got %q", out)
		}
	})

	t.Run("go-template", func(t *testing.T) {
		var buf bytes.Buffer
		f, err := ParseFormat("go-template={{ range .items }}{{ .name }} {{ end }}")
		if err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if err := Render(&buf, f, result); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.TrimSpace(buf.String()) != "x y" {
			t.Fatalf("got %q", buf.String())
		}
	})

	t.Run("slice of structs with custom-columns", func(t *testing.T) {
		results := []testResult{
			{ID: "abc", Count: 3, Items: []testItem{{Name: "x"}}},
			{ID: "def", Count: 5, Items: []testItem{{Name: "y"}}},
		}
		var buf bytes.Buffer
		f, err := ParseFormat("custom-columns=ID:.id,COUNT:.count")
		if err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if err := Render(&buf, f, results); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "abc") || !strings.Contains(out, "def") ||
			!strings.Contains(out, "3") || !strings.Contains(out, "5") {
			t.Fatalf("got %q", out)
		}
	})
}
