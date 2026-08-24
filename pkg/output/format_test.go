// pkg/output/format_test.go
package output

import "testing"

func TestParseFormat(t *testing.T) {
	cases := []struct {
		name    string
		spec    string
		wantErr bool
		want    FormatType
	}{
		{name: "json", spec: "json", want: FormatJSON},
		{name: "yaml", spec: "yaml", want: FormatYAML},
		{name: "plain", spec: "plain", want: FormatPlain},
		{name: "text alias", spec: "text", want: FormatPlain},
		{name: "empty is invalid", spec: "", wantErr: true},
		{name: "garbage is invalid", spec: "bogus", wantErr: true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f, err := ParseFormat(c.spec)
			if c.wantErr {
				if err == nil {
					t.Fatalf("expected error for spec %q, got nil", c.spec)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for spec %q: %v", c.spec, err)
			}
			if f.Type != c.want {
				t.Fatalf("spec %q: got type %q, want %q", c.spec, f.Type, c.want)
			}
			if f.RawSpec != c.spec {
				t.Fatalf("spec %q: got RawSpec %q, want %q", c.spec, f.RawSpec, c.spec)
			}
		})
	}
}
