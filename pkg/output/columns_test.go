package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseColumnsSpec(t *testing.T) {
	cases := []struct {
		name    string
		spec    string
		wantErr bool
		want    []ColumnSpec
	}{
		{
			name: "single column",
			spec: "NAME:.name",
			want: []ColumnSpec{{Header: "NAME", Path: ".name"}},
		},
		{
			name: "multiple columns",
			spec: "NAME:.name,STATUS:.status.phase",
			want: []ColumnSpec{
				{Header: "NAME", Path: ".name"},
				{Header: "STATUS", Path: ".status.phase"},
			},
		},
		{name: "missing colon", spec: "NAME.name", wantErr: true},
		{name: "empty header", spec: ":.name", wantErr: true},
		{name: "empty path", spec: "NAME:", wantErr: true},
		{name: "empty path after trim", spec: "NAME: ", wantErr: true},
		{
			name: "whitespace after comma is trimmed",
			spec: "NAME:.name, STATUS:.status",
			want: []ColumnSpec{
				{Header: "NAME", Path: ".name"},
				{Header: "STATUS", Path: ".status"},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := parseColumnsSpec(c.spec)
			if c.wantErr {
				if err == nil {
					t.Fatalf("expected error for spec %q, got nil", c.spec)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(c.want) {
				t.Fatalf("got %d columns, want %d", len(got), len(c.want))
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Fatalf("column %d: got %+v, want %+v", i, got[i], c.want[i])
				}
			}
		})
	}

	t.Run("large integral float64 renders without exponent notation", func(t *testing.T) {
		countColumns := []ColumnSpec{
			{Header: "COUNT", Path: ".count"},
		}
		data := map[string]any{"count": float64(1000000)}
		var buf bytes.Buffer
		if err := renderCustomColumns(&buf, countColumns, data); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "1000000") {
			t.Fatalf("expected output to contain 1000000, got: %s", out)
		}
		if strings.Contains(out, "e+") {
			t.Fatalf("expected no exponent notation, got: %s", out)
		}
	})
}

func TestRenderCustomColumns(t *testing.T) {
	columns := []ColumnSpec{
		{Header: "NAME", Path: ".name"},
		{Header: "STATUS", Path: ".status.phase"},
		{Header: "MISSING", Path: ".nope"},
	}

	t.Run("slice of objects", func(t *testing.T) {
		data := []map[string]any{
			{"name": "foo", "status": map[string]any{"phase": "active"}},
			{"name": "bar", "status": map[string]any{"phase": "pending"}},
		}
		var buf bytes.Buffer
		if err := renderCustomColumns(&buf, columns, data); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "NAME") || !strings.Contains(out, "STATUS") || !strings.Contains(out, "MISSING") {
			t.Fatalf("missing header in output: %s", out)
		}
		if !strings.Contains(out, "foo") || !strings.Contains(out, "active") {
			t.Fatalf("missing row data in output: %s", out)
		}
		if !strings.Contains(out, "<none>") {
			t.Fatalf("expected <none> for missing path, got: %s", out)
		}
	})

	t.Run("single object is wrapped into one row", func(t *testing.T) {
		data := map[string]any{"name": "foo", "status": map[string]any{"phase": "active"}}
		var buf bytes.Buffer
		if err := renderCustomColumns(&buf, columns, data); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "foo") || !strings.Contains(out, "active") {
			t.Fatalf("expected single object rendered as one row, got: %s", out)
		}
	})

	t.Run("unsupported type errors", func(t *testing.T) {
		var buf bytes.Buffer
		err := renderCustomColumns(&buf, columns, 42)
		if err == nil {
			t.Fatalf("expected error for unsupported type")
		}
	})

	t.Run("array-index dot-paths", func(t *testing.T) {
		itemsColumns := []ColumnSpec{
			{Header: "FIRST", Path: ".items[0].name"},
			{Header: "OUT_OF_RANGE", Path: ".items[5].name"},
			{Header: "MALFORMED_X", Path: ".items[x].name"},
			{Header: "MALFORMED_EMPTY", Path: ".items[].name"},
		}
		data := map[string]any{
			"items": []map[string]any{
				{"name": "foo"},
				{"name": "bar"},
			},
		}
		var buf bytes.Buffer
		if err := renderCustomColumns(&buf, itemsColumns, data); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "foo") {
			t.Fatalf("expected valid index .items[0].name to resolve to foo, got: %s", out)
		}
		// Out-of-range and malformed indexes must fall back to <none>, not
		// error or crash. There should be exactly 3 occurrences: one per
		// failing column (out-of-range, malformed "x", malformed empty).
		if count := strings.Count(out, "<none>"); count != 3 {
			t.Fatalf("expected 3 <none> occurrences (out-of-range + 2 malformed), got %d in: %s", count, out)
		}
	})
}
