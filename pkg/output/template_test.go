package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseGoTemplate(t *testing.T) {
	t.Run("valid template parses", func(t *testing.T) {
		tmpl, err := parseGoTemplate("{{ .name }}")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tmpl == nil {
			t.Fatalf("expected non-nil template")
		}
	})

	t.Run("syntax error caught at parse time", func(t *testing.T) {
		_, err := parseGoTemplate("{{ .name ")
		if err == nil {
			t.Fatalf("expected syntax error, got nil")
		}
	})
}

func TestRenderGoTemplate(t *testing.T) {
	t.Run("basic field access", func(t *testing.T) {
		tmpl, err := parseGoTemplate("{{ .name }}")
		if err != nil {
			t.Fatalf("parse error: %v", err)
		}
		var buf bytes.Buffer
		if err := renderGoTemplate(&buf, tmpl, map[string]any{"name": "foo"}); err != nil {
			t.Fatalf("render error: %v", err)
		}
		if buf.String() != "foo\n" {
			t.Fatalf("got %q, want %q", buf.String(), "foo\n")
		}
	})

	t.Run("helper funcs", func(t *testing.T) {
		cases := []struct {
			name string
			tmpl string
			data any
			want string
		}{
			{name: "lower", tmpl: "{{ lower .name }}", data: map[string]any{"name": "FOO"}, want: "foo\n"},
			{name: "upper", tmpl: "{{ upper .name }}", data: map[string]any{"name": "foo"}, want: "FOO\n"},
			{name: "join", tmpl: `{{ join .items "," }}`, data: map[string]any{"items": []string{"a", "b"}}, want: "a,b\n"},
			{name: "json", tmpl: "{{ json .name }}", data: map[string]any{"name": "foo"}, want: "\"foo\"\n"},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				tmpl, err := parseGoTemplate(c.tmpl)
				if err != nil {
					t.Fatalf("parse error: %v", err)
				}
				var buf bytes.Buffer
				if err := renderGoTemplate(&buf, tmpl, c.data); err != nil {
					t.Fatalf("render error: %v", err)
				}
				if buf.String() != c.want {
					t.Fatalf("got %q, want %q", buf.String(), c.want)
				}
			})
		}
	})

	t.Run("trailing newline is added when template omits it", func(t *testing.T) {
		tmpl, err := parseGoTemplate("{{ .name }}")
		if err != nil {
			t.Fatalf("parse error: %v", err)
		}
		var buf bytes.Buffer
		if err := renderGoTemplate(&buf, tmpl, map[string]any{"name": "foo"}); err != nil {
			t.Fatalf("render error: %v", err)
		}
		if !strings.HasSuffix(buf.String(), "\n") {
			t.Fatalf("expected trailing newline, got %q", buf.String())
		}
	})

	t.Run("trailing newline is not duplicated when template provides one", func(t *testing.T) {
		tmpl, err := parseGoTemplate("{{ .name }}\n")
		if err != nil {
			t.Fatalf("parse error: %v", err)
		}
		var buf bytes.Buffer
		if err := renderGoTemplate(&buf, tmpl, map[string]any{"name": "foo"}); err != nil {
			t.Fatalf("render error: %v", err)
		}
		if buf.String() != "foo\n" {
			t.Fatalf("got %q, want %q", buf.String(), "foo\n")
		}
	})
}
