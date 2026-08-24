package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestResolve(t *testing.T) {
	t.Run("empty spec defaults to plain", func(t *testing.T) {
		f, interactive, err := Resolve("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if interactive {
			t.Fatalf("expected non-interactive")
		}
		if f.Type != FormatPlain {
			t.Fatalf("got %v, want FormatPlain", f.Type)
		}
	})

	t.Run("explicit interactive spec resolves to interactive", func(t *testing.T) {
		_, interactive, err := Resolve("interactive")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !interactive {
			t.Fatalf("expected interactive")
		}
	})

	t.Run("explicit json spec resolves to non-interactive json", func(t *testing.T) {
		f, interactive, err := Resolve("json")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if interactive {
			t.Fatalf("expected non-interactive when --output is explicit")
		}
		if f.Type != FormatJSON {
			t.Fatalf("got %v, want FormatJSON", f.Type)
		}
	})

	t.Run("invalid explicit spec errors", func(t *testing.T) {
		_, _, err := Resolve("bogus")
		if err == nil {
			t.Fatalf("expected error for invalid spec")
		}
	})
}

func TestSetContextFromContext(t *testing.T) {
	cmd := &cobra.Command{}
	f, _ := ParseFormat("json")
	SetContext(cmd, f, false)

	got, interactive, err := FromContext(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Type != FormatJSON || interactive {
		t.Fatalf("got %v/%v, want json/false", got.Type, interactive)
	}
}

func TestFromContextMissing(t *testing.T) {
	cmd := &cobra.Command{}
	_, _, err := FromContext(cmd)
	if err == nil {
		t.Fatalf("expected error when format context is missing")
	}
}

func TestRequireFormat(t *testing.T) {
	jsonFmt, _ := ParseFormat("json")
	if err := RequireFormat(jsonFmt, FormatJSON, FormatYAML, FormatPlain); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	colsFmt, _ := ParseFormat("custom-columns=NAME:.name")
	err := RequireFormat(colsFmt, FormatJSON, FormatYAML, FormatPlain)
	if err == nil {
		t.Fatalf("expected error for disallowed format")
	}
	want := `format "custom-columns" is not supported for this command (supported formats: json, yaml, plain)`
	if err.Error() != want {
		t.Fatalf("got %q, want %q", err.Error(), want)
	}
}

func TestEmitNonInteractive(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	f, _ := ParseFormat("json")
	SetContext(cmd, f, false)

	if err := Emit(cmd, map[string]any{"name": "foo"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), `"name": "foo"`) {
		t.Fatalf("got %q", buf.String())
	}
}

func TestPlain_UsesFallbackForDefaultPlain(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	f, _ := ParseFormat("plain")
	SetContext(cmd, f, false)

	called := false
	err := Plain(cmd, map[string]any{"name": "foo"}, func() error {
		called = true
		cmd.Println("raw fallback output")
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("expected plainFallback to be called")
	}
	if !strings.Contains(buf.String(), "raw fallback output") {
		t.Fatalf("got %q", buf.String())
	}
}

func TestPlain_UsesEmitForNonPlainFormat(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	f, _ := ParseFormat("json")
	SetContext(cmd, f, false)

	err := Plain(cmd, map[string]any{"name": "foo"}, func() error {
		t.Fatalf("plainFallback should not be called for json format")
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), `"name": "foo"`) {
		t.Fatalf("got %q", buf.String())
	}
}

func TestPlainJSON(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := PlainJSON(cmd, map[string]any{"name": "foo"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := buf.String(); got != "{\n  \"name\": \"foo\"\n}\n" {
		t.Fatalf("got %q", got)
	}
}

func TestPlainJSONDoesNotEscapeHTML(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := PlainJSON(cmd, map[string]any{"name": "A&B <C> D"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := buf.String(); got != "{\n  \"name\": \"A&B <C> D\"\n}\n" {
		t.Fatalf("got %q", got)
	}
}

func TestPlainBare(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := PlainBare(cmd, "raw-value"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := buf.String(); got != "raw-value\n" {
		t.Fatalf("got %q", got)
	}
}

func TestPlainOneLine(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := PlainOneLine(cmd, []string{"a", "b"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := buf.String(); got != "a\nb\n" {
		t.Fatalf("got %q", got)
	}
}

func TestPlainJSONOrUnixUsesPlainJSONForDefaultPlain(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	f, _ := ParseFormat("plain")
	SetContext(cmd, f, false)

	if err := PlainJSONOrUnix(cmd, map[string]any{"name": "foo"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := buf.String(); got != "{\n  \"name\": \"foo\"\n}\n" {
		t.Fatalf("got %q", got)
	}
}

func TestPlainJSONOrUnixUsesPlainUnixForExplicitPlain(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	f, _ := ParseFormat("plain")
	SetContext(cmd, f, false)
	SetExplicit(cmd, true)

	if err := PlainJSONOrUnix(cmd, map[string]any{"name": "foo"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := buf.String(); got != "name=foo\n" {
		t.Fatalf("got %q", got)
	}
}

func TestPlainJSONOrUnixUsesPlainJSONWhenExplicitNotSet(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	f, _ := ParseFormat("plain")
	SetContext(cmd, f, false)

	if err := PlainJSONOrUnix(cmd, map[string]any{"name": "foo"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := buf.String(); got != "{\n  \"name\": \"foo\"\n}\n" {
		t.Fatalf("got %q", got)
	}
}

func TestPlainJSONOrUnixUsesEmitForNonPlainFormat(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	f, _ := ParseFormat("yaml")
	SetContext(cmd, f, false)

	if err := PlainJSONOrUnix(cmd, map[string]any{"name": "foo"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "name: foo") {
		t.Fatalf("got %q", buf.String())
	}
}

func TestPlainOrUnixUsesFallbackForDefaultPlain(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	f, _ := ParseFormat("plain")
	SetContext(cmd, f, false)

	err := PlainOrUnix(cmd, map[string]any{"name": "foo"}, func() error {
		cmd.Println("custom fallback")
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := buf.String(); got != "custom fallback\n" {
		t.Fatalf("got %q, want the custom fallback body, not the generic PlainJSON one", got)
	}
}

func TestPlainOrUnixUsesPlainUnixForExplicitPlain(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	f, _ := ParseFormat("plain")
	SetContext(cmd, f, false)
	SetExplicit(cmd, true)

	fallbackCalled := false
	err := PlainOrUnix(cmd, map[string]any{"name": "foo"}, func() error {
		fallbackCalled = true
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fallbackCalled {
		t.Fatalf("expected explicit --output plain to use PlainUnix, not the fallback")
	}
	if got := buf.String(); got != "name=foo\n" {
		t.Fatalf("got %q", got)
	}
}

func TestPlainOrUnixUsesEmitForNonPlainFormat(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	f, _ := ParseFormat("yaml")
	SetContext(cmd, f, false)

	fallbackCalled := false
	err := PlainOrUnix(cmd, map[string]any{"name": "foo"}, func() error {
		fallbackCalled = true
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fallbackCalled {
		t.Fatalf("expected --output yaml to use Emit, not the fallback")
	}
	if !strings.Contains(buf.String(), "name: foo") {
		t.Fatalf("got %q", buf.String())
	}
}

func TestEmitInteractiveFallsBackWithoutRealTerminal(t *testing.T) {
	orig := isRealTerminal
	isRealTerminal = func() bool { return false }
	t.Cleanup(func() { isRealTerminal = orig })

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	SetContext(cmd, Format{}, true)

	if err := Emit(cmd, map[string]any{"status": "set"}); err != nil {
		t.Fatalf("expected a graceful fallback instead of trying to open a TTY, got error: %v", err)
	}
	if !strings.Contains(buf.String(), `"status": "set"`) {
		t.Fatalf("expected the fallback to render the data, got %q", buf.String())
	}
}

func TestEmitInteractiveEmptyCollection(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	SetContext(cmd, Format{}, true)

	if err := Emit(cmd, []any{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "No resources found.") {
		t.Fatalf("got %q", buf.String())
	}
}

func TestEmitInteractiveNilCollection(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	SetContext(cmd, Format{}, true)

	var members []interface{}
	if err := Emit(cmd, members); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "No resources found.") {
		t.Fatalf("got %q", buf.String())
	}
}

func TestNormalizeForInteractiveWrapsScalarRows(t *testing.T) {
	rows, isSingleObject, err := normalizeForInteractive([]any{"a", "b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isSingleObject {
		t.Fatalf("expected a collection, not a single object")
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	for i, want := range []string{"a", "b"} {
		m, ok := rows[i].(map[string]any)
		if !ok {
			t.Fatalf("row %d: expected map[string]any, got %T", i, rows[i])
		}
		if m["value"] != want {
			t.Fatalf("row %d: got %v, want value=%q", i, m, want)
		}
	}
}
