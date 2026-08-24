// pkg/output/interactive_test.go
package output

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestTableModelColumnsAndRows(t *testing.T) {
	rows := []any{
		map[string]any{"id": "host1", "owner": "alice"},
		map[string]any{"id": "host2", "owner": "bob"},
	}
	m := newTableModel(rows)

	view := m.View().Content
	if !strings.Contains(view, "host1") || !strings.Contains(view, "host2") {
		t.Fatalf("expected both rows in view, got %q", view)
	}
	if !strings.Contains(view, "ID") || !strings.Contains(view, "OWNER") {
		t.Fatalf("expected column headers in view, got %q", view)
	}
}

func TestTableModelRendersScalarRows(t *testing.T) {
	rows, _, err := normalizeForInteractive([]any{"a", "b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := newTableModel(rows)

	view := m.View().Content
	if !strings.Contains(view, "a") || !strings.Contains(view, "b") {
		t.Fatalf("expected scalar row values in view, got %q", view)
	}
	if !strings.Contains(view, "VALUE") {
		t.Fatalf("expected synthesized VALUE column header in view, got %q", view)
	}
}

func TestTableModelEnterOpensDetail(t *testing.T) {
	rows := []any{
		map[string]any{"id": "host1", "owner": "alice"},
	}
	m := newTableModel(rows)

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	tm := updated.(tableModel)
	if tm.detail == nil {
		t.Fatalf("expected detail view to open after Enter")
	}
	if !strings.Contains(tm.detail.View().Content, "alice") {
		t.Fatalf("expected detail view to show selected row, got %q", tm.detail.View().Content)
	}
}

func TestTableModelEscReturnsFromDetail(t *testing.T) {
	rows := []any{
		map[string]any{"id": "host1", "owner": "alice"},
	}
	m := newTableModel(rows)
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	tm := updated.(tableModel)

	updated, _ = tm.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	tm = updated.(tableModel)
	if tm.detail != nil {
		t.Fatalf("expected detail view to close after Esc")
	}
}

func typeText(t *testing.T, m tableModel, s string) tableModel {
	t.Helper()
	for _, r := range s {
		updated, _ := m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		m = updated.(tableModel)
	}
	return m
}

func TestTableModelSlashOpensFilterInput(t *testing.T) {
	rows := []any{map[string]any{"id": "host1"}}
	m := newTableModel(rows)

	updated, _ := m.Update(tea.KeyPressMsg{Code: '/'})
	tm := updated.(tableModel)

	if !tm.filtering {
		t.Fatalf("expected '/' to enter filter-editing mode")
	}
	if !strings.Contains(tm.View().Content, "/") {
		t.Fatalf("expected the filter prompt in the view, got %q", tm.View().Content)
	}
}

func TestTableModelFilterHidesNonMatchingRows(t *testing.T) {
	rows := []any{
		map[string]any{"id": "prod-host"},
		map[string]any{"id": "dev-host"},
	}
	m := newTableModel(rows)

	updated, _ := m.Update(tea.KeyPressMsg{Code: '/'})
	tm := updated.(tableModel)
	tm = typeText(t, tm, "prod")

	if len(tm.visible) != 1 {
		t.Fatalf("expected exactly one visible row while typing \"prod\", got %d", len(tm.visible))
	}
	view := tm.View().Content
	if !strings.Contains(view, "prod-host") || strings.Contains(view, "dev-host") {
		t.Fatalf("expected only the matching row rendered, got %q", view)
	}
}

func TestTableModelFilterMatchesAnyVisibleColumn(t *testing.T) {
	rows := []any{
		map[string]any{"id": "host1", "owner": "alice"},
		map[string]any{"id": "host2", "owner": "bob"},
	}
	m := newTableModel(rows)

	updated, _ := m.Update(tea.KeyPressMsg{Code: '/'})
	tm := updated.(tableModel)
	tm = typeText(t, tm, "bob")

	if len(tm.visible) != 1 {
		t.Fatalf("expected the filter to match on a non-first column, got %d visible rows", len(tm.visible))
	}
}

func TestTableModelFilterIsCaseInsensitive(t *testing.T) {
	rows := []any{map[string]any{"id": "PROD-host"}, map[string]any{"id": "dev-host"}}
	m := newTableModel(rows)

	updated, _ := m.Update(tea.KeyPressMsg{Code: '/'})
	tm := updated.(tableModel)
	tm = typeText(t, tm, "prod")

	if len(tm.visible) != 1 {
		t.Fatalf("expected a case-insensitive match, got %d visible rows", len(tm.visible))
	}
}

func TestTableModelFilterEnterConfirmsAndShowsStatus(t *testing.T) {
	rows := []any{
		map[string]any{"id": "prod-host"},
		map[string]any{"id": "dev-host"},
	}
	m := newTableModel(rows)
	updated, _ := m.Update(tea.KeyPressMsg{Code: '/'})
	tm := updated.(tableModel)
	tm = typeText(t, tm, "prod")

	updated, _ = tm.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	tm = updated.(tableModel)

	if tm.filtering {
		t.Fatalf("expected Enter to exit filter-editing mode")
	}
	if tm.filterQuery != "prod" {
		t.Fatalf("expected the applied filter query to be %q, got %q", "prod", tm.filterQuery)
	}
	view := tm.View().Content
	if !strings.Contains(view, `"prod"`) || !strings.Contains(view, "1/2") {
		t.Fatalf("expected a filter status line naming the query and match count, got %q", view)
	}
}

func TestTableModelFilterEscWhileEditingRevertsToLastApplied(t *testing.T) {
	rows := []any{
		map[string]any{"id": "prod-host"},
		map[string]any{"id": "dev-host"},
	}
	m := newTableModel(rows)

	updated, _ := m.Update(tea.KeyPressMsg{Code: '/'})
	tm := updated.(tableModel)
	tm = typeText(t, tm, "prod")
	updated, _ = tm.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	tm = updated.(tableModel)

	updated, _ = tm.Update(tea.KeyPressMsg{Code: '/'})
	tm = updated.(tableModel)
	tm = typeText(t, tm, "xyz")
	if len(tm.visible) != 0 {
		t.Fatalf("expected the live edit to filter down to zero matches, got %d", len(tm.visible))
	}

	updated, _ = tm.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	tm = updated.(tableModel)

	if tm.filtering {
		t.Fatalf("expected Esc to exit filter-editing mode")
	}
	if tm.filterQuery != "prod" {
		t.Fatalf("expected Esc to revert to the last applied query %q, got %q", "prod", tm.filterQuery)
	}
	if len(tm.visible) != 1 {
		t.Fatalf("expected the reverted filter to restore the previous match count, got %d", len(tm.visible))
	}
}

func TestTableModelFilterEscClearsAppliedFilter(t *testing.T) {
	rows := []any{
		map[string]any{"id": "prod-host"},
		map[string]any{"id": "dev-host"},
	}
	m := newTableModel(rows)
	updated, _ := m.Update(tea.KeyPressMsg{Code: '/'})
	tm := updated.(tableModel)
	tm = typeText(t, tm, "prod")
	updated, _ = tm.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	tm = updated.(tableModel)

	updated, _ = tm.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	tm = updated.(tableModel)

	if tm.filterQuery != "" {
		t.Fatalf("expected Esc outside edit mode to clear the applied filter, got %q", tm.filterQuery)
	}
	if len(tm.visible) != len(rows) {
		t.Fatalf("expected all rows visible again after clearing the filter, got %d", len(tm.visible))
	}
}

func TestTableModelEnterOpensDetailForFilteredRow(t *testing.T) {
	rows := []any{
		map[string]any{"id": "prod-host"},
		map[string]any{"id": "dev-host"},
	}
	m := newTableModel(rows)
	updated, _ := m.Update(tea.KeyPressMsg{Code: '/'})
	tm := updated.(tableModel)
	tm = typeText(t, tm, "dev")
	updated, _ = tm.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	tm = updated.(tableModel)

	updated, _ = tm.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	tm = updated.(tableModel)

	if tm.detail == nil {
		t.Fatalf("expected Enter to open a detail view for the filtered row")
	}
	if !strings.Contains(tm.detail.View().Content, "dev-host") {
		t.Fatalf("expected the detail view to show the filtered row, got %q", tm.detail.View().Content)
	}
}

func TestTableModelFooterHintMentionsFilter(t *testing.T) {
	rows := []any{map[string]any{"id": "host1"}}
	m := newTableModel(rows)

	if !strings.Contains(m.View().Content, "filter") {
		t.Fatalf("expected the base footer hint to mention the filter key, got %q", m.View().Content)
	}
}

func TestDetailModelRendersNestedValue(t *testing.T) {
	m := newDetailModel(map[string]any{
		"id":          "host1",
		"annotations": map[string]any{"env": "prod"},
	}, false)
	view := m.View().Content
	if !strings.Contains(view, "host1") || !strings.Contains(view, "env") || !strings.Contains(view, "prod") {
		t.Fatalf("expected nested fields in detail view, got %q", view)
	}
}

func TestStandaloneDetailModelQuitsOnEsc(t *testing.T) {
	m := newDetailModel(map[string]any{"id": "host1"}, true)

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if cmd == nil {
		t.Fatalf("expected a tea.Quit cmd for a standalone detail model on Esc, got nil")
	}
}

func TestStandaloneDetailModelQuitsOnQ(t *testing.T) {
	m := newDetailModel(map[string]any{"id": "host1"}, true)

	_, cmd := m.Update(tea.KeyPressMsg{Code: 'q'})
	if cmd == nil {
		t.Fatalf("expected a tea.Quit cmd for a standalone detail model on q, got nil")
	}
}

func TestChildDetailModelDoesNotQuitOnEsc(t *testing.T) {
	m := newDetailModel(map[string]any{"id": "host1"}, false)

	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if cmd != nil {
		t.Fatalf("expected no cmd for a child detail model on Esc, got %v", cmd)
	}
	dm := updated.(detailModel)
	if !dm.closed {
		t.Fatalf("expected child detail model to be marked closed")
	}
}

func TestTableModelViewHasFooterHint(t *testing.T) {
	// A single-column scalar table (e.g. `conjur list`) never has hidden
	// columns, so the footer must not depend on that condition.
	rows, _, err := normalizeForInteractive([]any{"a", "b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := newTableModel(rows)

	view := m.View().Content
	if !strings.Contains(view, "quit") {
		t.Fatalf("expected a navigation footer hint in table view, got %q", view)
	}
}

func TestTableModelPageDownScrollsByPage(t *testing.T) {
	orig := terminalSizeFunc
	terminalSizeFunc = func() (int, int, error) { return 80, 5, nil }
	t.Cleanup(func() { terminalSizeFunc = orig })

	rows := make([]any, 50)
	for i := range rows {
		rows[i] = map[string]any{"id": fmt.Sprintf("host%02d", i)}
	}
	m := newTableModel(rows)

	before := m.table.Cursor()
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyPgDown})
	tm := updated.(tableModel)
	if tm.table.Cursor() <= before {
		t.Fatalf("expected PgDown to advance the cursor by a page, got cursor %d (was %d)", tm.table.Cursor(), before)
	}
}

func TestTableModelHeightClampedToTerminalHeight(t *testing.T) {
	orig := terminalSizeFunc
	terminalSizeFunc = func() (int, int, error) { return 80, 10, nil }
	t.Cleanup(func() { terminalSizeFunc = orig })

	rows := make([]any, 50)
	for i := range rows {
		rows[i] = map[string]any{"id": fmt.Sprintf("host%d", i)}
	}
	m := newTableModel(rows)

	if got := m.table.Height(); got >= len(rows)+1 {
		t.Fatalf("expected table height clamped below row count+1 (%d), got %d", len(rows)+1, got)
	}
}

// TestTableModelFooterPinnedToBottom guards against the footer hint
// trailing right after the last row on a short result set: the table
// viewport must pad down to the terminal height (like the detail view's
// viewport does) so the footer always lands on the terminal's last line.
func TestTableModelFooterPinnedToBottom(t *testing.T) {
	orig := terminalSizeFunc
	terminalSizeFunc = func() (int, int, error) { return 80, 10, nil }
	t.Cleanup(func() { terminalSizeFunc = orig })

	rows := []any{map[string]any{"id": "host1"}}
	m := newTableModel(rows)

	view := m.View().Content
	lines := strings.Split(view, "\n")
	if got := len(lines); got != 10 {
		t.Fatalf("expected the view to fill all 10 terminal lines so the footer lands on the last one, got %d lines:\n%s", got, view)
	}
	if last := lines[len(lines)-1]; last != tableFooterHint {
		t.Fatalf("expected footer hint on the last line, got %q", last)
	}
}

func TestTableModelHandlesWindowResize(t *testing.T) {
	rows := []any{
		map[string]any{"id": "host1", "owner": "alice"},
		map[string]any{"id": "host2", "owner": "bob"},
	}
	m := newTableModel(rows)

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 6})
	tm := updated.(tableModel)

	if got := tm.table.Width(); got != 40 {
		t.Fatalf("expected table width updated to 40, got %d", got)
	}
	if want, got := 4, tm.table.Height(); got != want {
		t.Fatalf("expected table to fill the new terminal height (want %d), got %d", want, got)
	}
}

func TestDetailModelHandlesWindowResize(t *testing.T) {
	m := newDetailModel(map[string]any{"id": "host1", "owner": "alice"}, false)

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	dm := updated.(detailModel)

	if got := dm.table.Width(); got != 40 {
		t.Fatalf("expected table width updated to 40, got %d", got)
	}
	if want, got := 8, dm.table.Height(); got != want {
		t.Fatalf("expected table to fill the new terminal height (want %d), got %d", want, got)
	}
}

func TestSingleValueDetailModelUsesViewport(t *testing.T) {
	m := newDetailModel(map[string]any{"id": "host1"}, false)

	if !m.singleValue {
		t.Fatalf("expected a single-field row to render via the viewport")
	}
}

func TestSingleValueDetailModelHandlesWindowResize(t *testing.T) {
	m := newDetailModel(map[string]any{"id": "host1"}, false)

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	dm := updated.(detailModel)

	if got := dm.viewport.Width(); got != 40 {
		t.Fatalf("expected viewport width updated to 40, got %d", got)
	}
	if got := dm.viewport.Height(); got >= 10 {
		t.Fatalf("expected viewport height reserved for footer hint, got %d", got)
	}
}

func TestSingleValueDetailModelShowsFullValueUntruncated(t *testing.T) {
	longValue := strings.Repeat("CERTDATA", 50)
	m := newDetailModel(longValue, true)

	view := m.View().Content
	unwrapped := strings.ReplaceAll(view, "\n", "")
	if !strings.Contains(unwrapped, longValue) {
		t.Fatalf("expected long scalar value to render in full (soft-wrapped, not truncated), got %q", view)
	}
	if strings.Contains(view, "…") {
		t.Fatalf("expected no ellipsis truncation for a single value, got %q", view)
	}
}

func TestDetailModelViewHasScrollingFooter(t *testing.T) {
	m := newDetailModel(map[string]any{"id": "host1"}, false)

	view := m.View().Content
	if !strings.Contains(view, "scroll") {
		t.Fatalf("expected a scroll hint/footer in detail view, got %q", view)
	}
}

func TestDetailModelIsATable(t *testing.T) {
	m := newDetailModel(map[string]any{
		"id":          "host1",
		"annotations": map[string]any{"env": "prod"},
	}, false)

	view := m.View().Content
	if !strings.Contains(view, "FIELD") || !strings.Contains(view, "VALUE") {
		t.Fatalf("expected FIELD/VALUE column headers in detail view, got %q", view)
	}
	if !strings.Contains(view, "annotations.env") {
		t.Fatalf("expected dotted field path for nested value, got %q", view)
	}
}

func TestDetailModelStylesFieldColumnBold(t *testing.T) {
	m := newDetailModel(map[string]any{"id": "host1"}, false)

	view := m.View().Content
	if !strings.Contains(view, detailFieldStyle().Render("id")) {
		t.Fatalf("expected FIELD column to be rendered with detailFieldStyle, got %q", view)
	}
}

func TestDetailModelZebraStripesValueColumn(t *testing.T) {
	m := newDetailModel(map[string]any{"a": "first", "b": "second"}, false)

	view := m.View().Content
	if !strings.Contains(view, detailZebraValueStyle().Render("second")) {
		t.Fatalf("expected the second row's VALUE to be zebra-striped, got %q", view)
	}
	if strings.Contains(view, detailZebraValueStyle().Render("first")) {
		t.Fatalf("expected the first row's VALUE to be unstriped, got %q", view)
	}
}

func TestDetailModelGroupsArraysUnderHeaderRow(t *testing.T) {
	m := newDetailModel(map[string]any{"tags": []any{"a", "b", "c"}}, false)

	view := m.View().Content
	if !strings.Contains(view, "tags") || !strings.Contains(view, "3 items") {
		t.Fatalf("expected an array header row naming the field and its item count, got %q", view)
	}
	if !strings.Contains(view, "[0]") || !strings.Contains(view, "a") {
		t.Fatalf("expected indexed child rows under the header, got %q", view)
	}
	if strings.Contains(view, "tags[0]") {
		t.Fatalf("expected array items to drop the repeated field prefix once grouped under a header, got %q", view)
	}
}

func TestDetailModelPageDownScrollsByPage(t *testing.T) {
	orig := terminalSizeFunc
	terminalSizeFunc = func() (int, int, error) { return 80, 5, nil }
	t.Cleanup(func() { terminalSizeFunc = orig })

	row := map[string]any{}
	for i := 0; i < 50; i++ {
		row[fmt.Sprintf("field%02d", i)] = "value"
	}
	m := newDetailModel(row, false)

	before := m.table.Cursor()
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyPgDown})
	dm := updated.(detailModel)
	if dm.table.Cursor() <= before {
		t.Fatalf("expected PgDown to advance the cursor by a page, got cursor %d (was %d)", dm.table.Cursor(), before)
	}
}
