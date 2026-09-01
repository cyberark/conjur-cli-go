package output

import (
	"fmt"
	"sort"
	"strings"
)

const cellSummaryMaxItems = 4

// deriveColumns returns the ordered set of top-level string keys across all
// rows: "id" or "name" (whichever is present) pinned first, remaining keys
// alphabetical. Used to build interactive table columns automatically, since
// there is no per-command column spec in this design.
func deriveColumns(rows []any) []string {
	seen := make(map[string]struct{})
	for _, row := range rows {
		m, ok := row.(map[string]any)
		if !ok {
			continue
		}
		for k := range m {
			seen[k] = struct{}{}
		}
	}

	pinned := ""
	for _, candidate := range []string{"id", "name"} {
		if _, ok := seen[candidate]; ok {
			pinned = candidate
			break
		}
	}

	rest := make([]string, 0, len(seen))
	for k := range seen {
		if k == pinned {
			continue
		}
		rest = append(rest, k)
	}
	sort.Strings(rest)

	if pinned == "" {
		return rest
	}
	return append([]string{pinned}, rest...)
}

// cellValue stringifies row[key] for display in an interactive table cell.
// Scalars stringify directly; arrays/objects render as a truncated summary
// (full value is available via the detail-view drill-down); a missing key
// renders "<none>", matching custom-columns' evalPath convention.
func cellValue(row any, key string) string {
	m, ok := row.(map[string]any)
	if !ok {
		return "<none>"
	}
	v, ok := m[key]
	if !ok {
		return "<none>"
	}

	switch value := v.(type) {
	case []any:
		return summarizeSlice(value)
	case map[string]any:
		return summarizeMap(value)
	default:
		return evalPath(row, "."+key)
	}
}

func summarizeSlice(items []any) string {
	parts := make([]string, 0, len(items))
	for _, item := range items {
		parts = append(parts, fmt.Sprint(item))
	}
	return truncateJoined(parts)
}

func summarizeMap(m map[string]any) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", k, m[k]))
	}
	return truncateJoined(parts)
}

func truncateJoined(parts []string) string {
	if len(parts) <= cellSummaryMaxItems {
		return strings.Join(parts, ", ")
	}
	return strings.Join(parts[:cellSummaryMaxItems], ", ") + "…"
}
