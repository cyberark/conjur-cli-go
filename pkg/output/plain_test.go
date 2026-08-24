package output

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func TestPlainUnixScalar(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := PlainUnix(cmd, "raw-value"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := buf.String(); got != "raw-value\n" {
		t.Fatalf("got %q", got)
	}
}

func TestPlainUnixStringSlice(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := PlainUnix(cmd, []string{"a", "b"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := buf.String(); got != "a\nb\n" {
		t.Fatalf("got %q", got)
	}
}

func TestPlainUnixSingleObject(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := PlainUnix(cmd, map[string]any{"name": "foo", "id": 1}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := buf.String(); got != "id=1\nname=foo\n" {
		t.Fatalf("got %q", got)
	}
}

func TestPlainUnixListOfObjects(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	data := []map[string]any{
		{"name": "a", "id": 1},
		{"name": "b", "id": 2},
	}
	if err := PlainUnix(cmd, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := buf.String(); got != "1:a\n2:b\n" {
		t.Fatalf("got %q", got)
	}
}

func TestPlainUnixDoesNotEscapeHTML(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := PlainUnix(cmd, map[string]any{"name": "A&B <C> D"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := buf.String(); got != "name=A&B <C> D\n" {
		t.Fatalf("got %q", got)
	}
}
