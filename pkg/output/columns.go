package output

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"text/tabwriter"
)

func parseColumnsSpec(spec string) ([]ColumnSpec, error) {
	pairs := strings.Split(spec, ",")
	cols := make([]ColumnSpec, 0, len(pairs))
	for _, pair := range pairs {
		idx := strings.Index(pair, ":")
		if idx < 0 {
			return nil, fmt.Errorf("invalid custom-columns pair %q: missing ':'", pair)
		}
		header := strings.TrimSpace(pair[:idx])
		path := strings.TrimSpace(pair[idx+1:])
		if header == "" {
			return nil, fmt.Errorf("invalid custom-columns pair %q: empty header", pair)
		}
		if path == "" {
			return nil, fmt.Errorf("invalid custom-columns pair %q: empty path", pair)
		}
		cols = append(cols, ColumnSpec{Header: header, Path: path})
	}
	return cols, nil
}

func normalizeRows(data any) ([]any, error) {
	// Round-trip through JSON so callers can pass any marshalable struct.
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("custom-columns: marshaling data: %w", err)
	}
	var unmarshaled any
	if err := json.Unmarshal(raw, &unmarshaled); err != nil {
		return nil, fmt.Errorf("custom-columns: unmarshaling data: %w", err)
	}

	switch v := unmarshaled.(type) {
	case []any:
		return v, nil
	case map[string]any:
		return []any{v}, nil
	default:
		return nil, fmt.Errorf("custom-columns requires a struct, map, or slice, got %T", unmarshaled)
	}
}

// evalPath resolves a dot-path like ".status.phase" or ".items[0].name"
// against a JSON-round-tripped value. Returns "<none>" if the path
// doesn't resolve.
func evalPath(row any, path string) string {
	path = strings.TrimPrefix(path, ".")
	if path == "" {
		return "<none>"
	}
	segments := strings.Split(path, ".")
	current := row
	for _, seg := range segments {
		field, index, hasIndex := splitIndex(seg)
		m, ok := current.(map[string]any)
		if !ok {
			return "<none>"
		}
		next, ok := m[field]
		if !ok {
			return "<none>"
		}
		if hasIndex {
			arr, ok := next.([]any)
			if !ok || index < 0 || index >= len(arr) {
				return "<none>"
			}
			current = arr[index]
		} else {
			current = next
		}
	}
	if f, ok := current.(float64); ok {
		if f == math.Trunc(f) && !math.IsInf(f, 0) {
			return strconv.FormatInt(int64(f), 10)
		}
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
	return fmt.Sprintf("%v", current)
}

// splitIndex splits "items[0]" into ("items", 0, true), or returns
// (seg, 0, false) if seg has no index suffix.
func splitIndex(seg string) (string, int, bool) {
	open := strings.Index(seg, "[")
	if open < 0 || !strings.HasSuffix(seg, "]") {
		return seg, 0, false
	}
	field := seg[:open]
	idxStr := seg[open+1 : len(seg)-1]
	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		return seg, 0, false
	}
	return field, idx, true
}

func renderCustomColumns(w io.Writer, columns []ColumnSpec, data any) error {
	rows, err := normalizeRows(data)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 8, 2, ' ', 0)

	headers := make([]string, len(columns))
	for i, c := range columns {
		headers[i] = c.Header
	}
	fmt.Fprintln(tw, strings.Join(headers, "\t"))

	for _, row := range rows {
		cells := make([]string, len(columns))
		for i, c := range columns {
			cells[i] = evalPath(row, c.Path)
		}
		fmt.Fprintln(tw, strings.Join(cells, "\t"))
	}

	return tw.Flush()
}
