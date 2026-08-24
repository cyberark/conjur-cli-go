package output

import "testing"

func TestDeriveColumns(t *testing.T) {
	cases := []struct {
		name string
		rows []any
		want []string
	}{
		{
			name: "pins id first",
			rows: []any{
				map[string]any{"zeta": "z", "id": "1", "alpha": "a"},
			},
			want: []string{"id", "alpha", "zeta"},
		},
		{
			name: "pins name first when no id",
			rows: []any{
				map[string]any{"zeta": "z", "name": "n"},
			},
			want: []string{"name", "zeta"},
		},
		{
			name: "union of keys across rows, alphabetical after pinned",
			rows: []any{
				map[string]any{"id": "1", "b": "x"},
				map[string]any{"id": "2", "a": "y"},
			},
			want: []string{"id", "a", "b"},
		},
		{
			name: "empty rows",
			rows: []any{},
			want: []string{},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := deriveColumns(c.rows)
			if len(got) != len(c.want) {
				t.Fatalf("got %v, want %v", got, c.want)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Fatalf("got %v, want %v", got, c.want)
				}
			}
		})
	}
}

func TestCellValue(t *testing.T) {
	cases := []struct {
		name string
		row  any
		key  string
		want string
	}{
		{
			name: "scalar string",
			row:  map[string]any{"id": "foo"},
			key:  "id",
			want: "foo",
		},
		{
			name: "missing key renders none",
			row:  map[string]any{"id": "foo"},
			key:  "owner",
			want: "<none>",
		},
		{
			name: "array of scalars joins and truncates",
			row:  map[string]any{"roles": []any{"admin", "ops", "dev", "qa", "infra"}},
			key:  "roles",
			want: "admin, ops, dev, qa…",
		},
		{
			name: "short array of scalars joins without truncation",
			row:  map[string]any{"roles": []any{"admin", "ops"}},
			key:  "roles",
			want: "admin, ops",
		},
		{
			name: "object renders key=val pairs truncated",
			row:  map[string]any{"annotations": map[string]any{"env": "prod", "team": "wolves", "owner": "sre", "tier": "1", "extra": "x"}},
			key:  "annotations",
			want: "env=prod, extra=x, owner=sre, team=wolves…",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := cellValue(c.row, c.key)
			if got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
}
