package output

import (
	"bytes"
	"testing"
)

func TestRenderJSON(t *testing.T) {
	data := map[string]any{"name": "foo", "count": 2}
	var buf bytes.Buffer
	if err := renderJSON(&buf, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "{\n  \"count\": 2,\n  \"name\": \"foo\"\n}\n"
	if buf.String() != want {
		t.Fatalf("got %q, want %q", buf.String(), want)
	}
}

func TestRenderJSONDoesNotEscapeHTML(t *testing.T) {
	data := map[string]any{"name": "A&B <C> D"}
	var buf bytes.Buffer
	if err := renderJSON(&buf, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "{\n  \"name\": \"A&B <C> D\"\n}\n"
	if buf.String() != want {
		t.Fatalf("got %q, want %q", buf.String(), want)
	}
}
