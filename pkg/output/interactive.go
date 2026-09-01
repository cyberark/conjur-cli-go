// pkg/output/interactive.go
package output

import (
	"fmt"
	"image/color"
	"os"
	"sort"
	"strings"
	"sync"

	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const minColumnWidth = 8
const fallbackTerminalWidth = 80
const fallbackTerminalHeight = 24

// detailFooterHint documents the detail view's scroll/close keys, mirroring
// the table view's footer so both views explain themselves.
const detailFooterHint = "↑/↓ scroll · PgUp/PgDn (b/f) page · q/Esc back"

// tableFooterHint documents the table view's navigation keys. Shown
// unconditionally — unlike the hidden-columns hint, which only appears when
// fitColumns actually dropped columns — so single-column tables (e.g.
// `conjur list`) still explain how to drill in or quit.
const tableFooterHint = "↑/↓ navigate · PgUp/PgDn (b/f) page · Enter detail · / filter · q quit"

type tableModel struct {
	table   table.Model
	rows    []any    // full, unfiltered row set
	visible []any    // rows currently shown, after applying filterQuery (== rows when no filter)
	columns []string // fitted column names, in table order; used to resize without re-deriving/dropping columns
	hidden  []string // column names dropped for width, shown in the footer hint
	detail  *detailModel

	filterInput textinput.Model
	filtering   bool   // true while the filter input has focus and is accepting keystrokes
	filterQuery string // last applied filter; empty means no filter is active
}

func newTableModel(rows []any) tableModel {
	allColumns := deriveColumns(rows)
	width := terminalWidth()

	columns, hidden := fitColumns(allColumns, width)

	t := table.New(
		table.WithColumns(buildColumns(columns, width)),
		table.WithRows(rowsToTableRows(rows, columns)),
		table.WithFocused(true),
		table.WithWidth(width),
		table.WithHeight(tableHeight(terminalHeight())),
	)

	input := textinput.New()
	input.Prompt = "/ "

	return tableModel{table: t, rows: rows, visible: rows, columns: columns, hidden: hidden, filterInput: input}
}

// rowsToTableRows renders rows into table.Rows for the given fitted columns,
// shared by the initial build and every filter re-application.
func rowsToTableRows(rows []any, columns []string) []table.Row {
	tblRows := make([]table.Row, len(rows))
	for i, row := range rows {
		cells := make(table.Row, len(columns))
		for j, name := range columns {
			cells[j] = cellValue(row, name)
		}
		tblRows[i] = cells
	}
	return tblRows
}

// filterRows keeps rows where at least one fitted column's value contains
// query, case-insensitively. An empty query matches everything.
func filterRows(rows []any, columns []string, query string) []any {
	if query == "" {
		return rows
	}
	q := strings.ToLower(query)
	out := make([]any, 0, len(rows))
	for _, row := range rows {
		for _, col := range columns {
			if strings.Contains(strings.ToLower(cellValue(row, col)), q) {
				out = append(out, row)
				break
			}
		}
	}
	return out
}

// previewFilter re-derives the visible row set for query and pushes it into
// the underlying table, resetting the cursor since the old position may no
// longer correspond to the same row. It does not touch filterQuery, so
// live keystrokes can preview a filter without disturbing the last applied
// value that Esc reverts to.
func (m *tableModel) previewFilter(query string) {
	m.visible = filterRows(m.rows, m.columns, query)
	m.table.SetRows(rowsToTableRows(m.visible, m.columns))
	m.table.SetCursor(0)
}

// applyFilter commits query as the applied filter (shown in the status
// line and restored the next time the filter box opens) and previews it.
func (m *tableModel) applyFilter(query string) {
	m.filterQuery = query
	m.previewFilter(query)
}

// buildColumns renders fitted column names into table.Columns, splitting
// width evenly across them.
func buildColumns(names []string, width int) []table.Column {
	cols := make([]table.Column, len(names))
	for i, name := range names {
		cols[i] = table.Column{Title: strings.ToUpper(name), Width: width / max(1, len(names))}
	}
	return cols
}

// tableHeight fills the terminal height, minus one line reserved for the
// footer hint — mirroring the detail view's viewport, which always fills
// down to the footer rather than stopping after the last row. The
// underlying table viewport pads short results with blank lines, so the
// footer hint stays pinned to the bottom of the screen instead of trailing
// right after the last record.
func tableHeight(terminalHeight int) int {
	limit := terminalHeight - 1
	if limit < 1 {
		limit = 1
	}
	return limit
}

// fitColumns drops lowest-priority (alphabetically-last, non-pinned) columns
// until the remainder fits at minColumnWidth per column within width.
func fitColumns(columns []string, width int) (fitted []string, dropped []string) {
	fitted = columns
	for len(fitted) > 1 && width/len(fitted) < minColumnWidth {
		dropped = append(dropped, fitted[len(fitted)-1])
		fitted = fitted[:len(fitted)-1]
	}
	sort.Strings(dropped)
	return fitted, dropped
}

func terminalWidth() int {
	w, _, err := terminalSizeFunc()
	if err != nil || w <= 0 {
		return fallbackTerminalWidth
	}
	return w
}

func terminalHeight() int {
	_, h, err := terminalSizeFunc()
	if err != nil || h <= 0 {
		return fallbackTerminalHeight
	}
	return h
}

func (m tableModel) Init() tea.Cmd {
	return nil
}

func (m tableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.detail != nil {
		updated, cmd := m.detail.Update(msg)
		dm := updated.(detailModel)
		if dm.closed {
			m.detail = nil
			return m, nil
		}
		m.detail = &dm
		return m, cmd
	}

	if m.filtering {
		return m.updateFiltering(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.table.SetColumns(buildColumns(m.columns, msg.Width))
		m.table.SetWidth(msg.Width)
		m.table.SetHeight(tableHeight(msg.Height))
		return m, nil
	case tea.KeyPressMsg:
		if msg.Mod == tea.ModCtrl && msg.Code == 'c' {
			return m, tea.Quit
		}
		switch msg.Code {
		case tea.KeyEnter:
			selected := m.table.SelectedRow()
			if selected != nil {
				idx := m.table.Cursor()
				if idx >= 0 && idx < len(m.visible) {
					dm := newDetailModel(m.visible[idx], false)
					m.detail = &dm
				}
			}
			return m, nil
		case '/':
			m.filtering = true
			m.filterInput.SetValue(m.filterQuery)
			m.filterInput.CursorEnd()
			return m, m.filterInput.Focus()
		case tea.KeyEsc:
			if m.filterQuery != "" {
				m.applyFilter("")
			}
			return m, nil
		case 'q':
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// updateFiltering handles input while the filter box has focus: every
// keystroke live-filters the table, Enter confirms and returns focus to the
// table, and Esc reverts to the last applied query (or clears it if none
// was ever applied) before returning focus.
func (m tableModel) updateFiltering(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.Code {
		case tea.KeyEnter:
			m.filtering = false
			m.applyFilter(m.filterInput.Value())
			return m, nil
		case tea.KeyEsc:
			m.filtering = false
			m.filterInput.SetValue(m.filterQuery)
			m.previewFilter(m.filterQuery)
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.filterInput, cmd = m.filterInput.Update(msg)
	m.previewFilter(m.filterInput.Value())
	return m, cmd
}

func (m tableModel) View() tea.View {
	if m.detail != nil {
		return m.detail.View()
	}
	view := m.table.View()
	if len(m.hidden) > 0 {
		view += fmt.Sprintf("\n+%d more fields — press Enter for full detail", len(m.hidden))
	}
	view += "\n" + m.footerLine()
	v := tea.NewView(view)
	v.AltScreen = true
	return v
}

// footerLine picks the table view's bottom line: the filter box itself
// while editing, a status line naming the applied query and match count
// once one is set, or the plain navigation hint otherwise.
func (m tableModel) footerLine() string {
	if m.filtering {
		return m.filterInput.View()
	}
	if m.filterQuery != "" {
		return fmt.Sprintf("Filter: %q (%d/%d) — / edit · Esc clear", m.filterQuery, len(m.visible), len(m.rows))
	}
	return tableFooterHint
}

type detailModel struct {
	row    any
	closed bool
	// standalone is true when this detailModel is the root tea.Model (Emit's
	// single-object branch), and false when a tableModel owns it as a child.
	// A child's parent observes closed and drops the child itself; a
	// standalone model has no parent to observe it, so it must return
	// tea.Quit itself on close or the program can never exit on q/Esc.
	standalone bool
	table      table.Model
	rows       []table.Row

	// singleValue is true when the row flattens to exactly one field (e.g.
	// `conjur variable value`'s scalar result, or a single-attribute
	// object). A one-row table wastes most of the screen truncating long
	// content with "…" instead of showing it, so single values render in a
	// soft-wrapping, scrollable viewport instead.
	singleValue bool
	viewport    viewport.Model
	valueRow    detailRow
}

// hasDarkBackground is queried lazily, on first use, and cached from then
// on — not at package init, since every style below is a package-level var
// and package-level var initializers run unconditionally at program
// startup for any command that imports this package. Deferring the actual
// query to first Render() call means it only ever runs for commands that
// reach an interactive view, matching the precondition
// lipgloss.HasDarkBackground itself relies on (both stdin and stdout being
// a real TTY) rather than firing (and, on Unix, blocking on a terminal
// escape-sequence round trip) for every command, interactive or not.
var hasDarkBackground = sync.OnceValue(func() bool {
	return lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
})

// adaptiveColor resolves to dark on a dark-background terminal and light
// otherwise, for the interactive views' accent colors: fixed ANSI-256 codes
// tuned for a dark terminal read poorly on light backgrounds (e.g. code
// 245's light gray is nearly invisible on white).
func adaptiveColor(light, dark string) color.Color {
	if hasDarkBackground() {
		return lipgloss.Color(dark)
	}
	return lipgloss.Color(light)
}

// adaptivePink etc. are functions, not resolved colors, for the same
// lazy-evaluation reason as hasDarkBackground above: a package-level
// var whose initializer calls adaptiveColor would resolve (and so query
// the terminal) at package init regardless of whether interactive mode
// ever runs.
var (
	adaptivePink  = sync.OnceValue(func() color.Color { return adaptiveColor("162", "212") })
	adaptiveGray  = sync.OnceValue(func() color.Color { return adaptiveColor("238", "245") })
	adaptiveGreen = sync.OnceValue(func() color.Color { return adaptiveColor("28", "78") })
	adaptiveAmber = sync.OnceValue(func() color.Color { return adaptiveColor("130", "214") })
)

// detailFieldStyle bolds and accent-colors the FIELD column so it reads as
// a label rather than data, echoing the bold labels the old text-dump detail
// view used before it became a table. A function, not a resolved Style,
// so building it doesn't resolve adaptivePink (and so query the terminal)
// until a detail view actually renders.
var detailFieldStyle = sync.OnceValue(func() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(adaptivePink())
})

// detailZebraValueStyle dims every other row's VALUE cell, giving a long
// flattened field list (many rows, one per leaf) a visual rhythm that's
// otherwise lost once nesting collapses into dotted paths.
var detailZebraValueStyle = sync.OnceValue(func() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(adaptiveGray())
})

// detailArrayCountStyle renders an array header row's item count, visually
// distinct from a leaf VALUE so it reads as metadata about the rows below it
// rather than data of its own.
var detailArrayCountStyle = sync.OnceValue(func() lipgloss.Style {
	return lipgloss.NewStyle().Italic(true).Foreground(adaptiveGray())
})

func newDetailModel(row any, standalone bool) detailModel {
	width := terminalWidth()
	height := terminalHeight()
	detailRows := flattenDetailRows(row)
	rawRows := toTableRows(detailRows)

	if len(detailRows) == 1 {
		return newSingleValueDetailModel(row, detailRows[0], standalone, width, height)
	}

	styles := table.DefaultStyles()
	styles.Header = styles.Header.Foreground(adaptivePink())

	t := table.New(
		table.WithColumns(detailColumns(rawRows, width)),
		table.WithRows(styleDetailRows(detailRows)),
		table.WithFocused(true),
		table.WithWidth(width),
		table.WithHeight(tableHeight(height)),
		table.WithStyles(styles),
	)

	return detailModel{row: row, standalone: standalone, table: t, rows: rawRows}
}

// newSingleValueDetailModel renders a one-field detail (e.g. a scalar
// secret value) in a soft-wrapping, scrollable viewport rather than a
// one-row table, so long or multi-line content (PEM certificates, JSON
// blobs) reads in full instead of being ellipsis-truncated to one line.
func newSingleValueDetailModel(row any, r detailRow, standalone bool, width, height int) detailModel {
	vp := viewport.New(viewport.WithWidth(width), viewport.WithHeight(detailViewportHeight(height)))
	vp.SoftWrap = true
	vp.SetContent(singleValueContent(r))

	return detailModel{row: row, standalone: standalone, singleValue: true, viewport: vp, valueRow: r}
}

// singleValueContent renders the FIELD label above the VALUE so the
// viewport reads like the table it replaces; an unnamed row (a bare
// scalar with no field path) skips the label since there's nothing to
// name.
func singleValueContent(r detailRow) string {
	if r.field == "" {
		return r.value
	}
	return detailFieldStyle().Render(r.field) + "\n\n" + r.value
}

// detailViewportHeight reserves one line for detailFooterHint, mirroring
// tableHeight's reservation for the table view's footer hint.
func detailViewportHeight(terminalHeight int) int {
	h := terminalHeight - 1
	if h < 1 {
		h = 1
	}
	return h
}

// toTableRows strips detailRow down to the plain [field, value] pair
// table.Row needs, discarding the header flag once styleDetailRows has had
// a chance to use it.
func toTableRows(rows []detailRow) []table.Row {
	out := make([]table.Row, len(rows))
	for i, r := range rows {
		out[i] = table.Row{r.field, r.value}
	}
	return out
}

// styleDetailRows renders each row's cells. Column widths (detailColumns)
// must be computed from the plain rows (toTableRows), not these — ANSI
// escape bytes baked in here would otherwise inflate the plain len() width
// calculation.
func styleDetailRows(rows []detailRow) []table.Row {
	styled := make([]table.Row, len(rows))
	for i, r := range rows {
		if r.header {
			styled[i] = table.Row{detailFieldStyle().Render(r.field), detailArrayCountStyle().Render(r.value)}
			continue
		}
		value := r.value
		if i%2 == 1 {
			value = detailZebraValueStyle().Render(value)
		}
		styled[i] = table.Row{detailFieldStyle().Render(r.field), value}
	}
	return styled
}

// detailRow is a pre-styling FIELD/VALUE pair. header marks an array's
// summary row (see appendDetailRows) so styleDetailRows can render it
// differently from a leaf row.
type detailRow struct {
	field  string
	value  string
	header bool
}

// flattenDetailRows walks row into FIELD/VALUE rows, one per leaf scalar,
// rather than the summarized single-cell form cellValue uses for the outer
// table — the detail view exists precisely to show what that summary
// truncated. Nested maps flatten into dotted paths (e.g. "annotations.env").
// Arrays get their own header row ("tags (3 items)") followed by indented
// index rows ("[0]", "[1]", ...) instead of repeating the array's field name
// on every item, and nested collections indent one level further under
// whichever array contains them.
func flattenDetailRows(row any) []detailRow {
	var rows []detailRow
	appendDetailRows(&rows, "", 0, row)
	return rows
}

func appendDetailRows(rows *[]detailRow, prefix string, depth int, v any) {
	switch value := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(value))
		for k := range value {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			appendDetailRows(rows, joinFieldPath(prefix, k), depth, value[k])
		}
	case []any:
		label := prefix
		if label == "" {
			label = "(array)"
		}
		*rows = append(*rows, detailRow{
			field:  detailIndent(depth) + label,
			value:  arrayCountSummary(len(value)),
			header: true,
		})
		for i, item := range value {
			appendDetailRows(rows, fmt.Sprintf("[%d]", i), depth+1, item)
		}
	default:
		*rows = append(*rows, detailRow{field: detailIndent(depth) + prefix, value: fmt.Sprintf("%v", value)})
	}
}

func detailIndent(depth int) string {
	return strings.Repeat("  ", depth)
}

func arrayCountSummary(n int) string {
	if n == 1 {
		return "(1 item)"
	}
	return fmt.Sprintf("(%d items)", n)
}

func joinFieldPath(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}

// detailColumns sizes FIELD to fit the longest field path (capped at half
// the terminal width) and gives the remainder to VALUE, since values —
// unlike field names — are the content worth reading in full.
func detailColumns(rows []table.Row, width int) []table.Column {
	fieldWidth := minColumnWidth
	for _, r := range rows {
		if l := len(r[0]) + 2; l > fieldWidth {
			fieldWidth = l
		}
	}
	if fieldWidth > width/2 {
		fieldWidth = width / 2
	}
	valueWidth := width - fieldWidth
	if valueWidth < 1 {
		valueWidth = 1
	}
	return []table.Column{
		{Title: "FIELD", Width: fieldWidth},
		{Title: "VALUE", Width: valueWidth},
	}
}

func (m detailModel) Init() tea.Cmd {
	return nil
}

func (m detailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if m.singleValue {
			m.viewport.SetWidth(msg.Width)
			m.viewport.SetHeight(detailViewportHeight(msg.Height))
			return m, nil
		}
		m.table.SetColumns(detailColumns(m.rows, msg.Width))
		m.table.SetWidth(msg.Width)
		m.table.SetHeight(tableHeight(msg.Height))
		return m, nil
	case tea.KeyPressMsg:
		if msg.Mod == tea.ModCtrl && msg.Code == 'c' {
			return m, tea.Quit
		}
		switch msg.Code {
		case tea.KeyEsc, 'q':
			m.closed = true
			if m.standalone {
				return m, tea.Quit
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	if m.singleValue {
		m.viewport, cmd = m.viewport.Update(msg)
	} else {
		m.table, cmd = m.table.Update(msg)
	}
	return m, cmd
}

func (m detailModel) View() tea.View {
	body := m.table.View()
	if m.singleValue {
		body = m.viewport.View()
	}
	view := body + "\n" + detailFooterHint
	v := tea.NewView(view)
	v.AltScreen = true
	return v
}
