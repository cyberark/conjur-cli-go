package output

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"charm.land/lipgloss/v2"
)

// renderTable prints data as a static (non-interactive) tabwriter-aligned
// table: column derivation and cell formatting are shared with the
// interactive table view (deriveColumns/cellValue in interactive_columns.go)
// so both views agree on shape. No column is ever dropped (unlike the
// interactive view), but each cell is soft-wrapped to its column's share of
// the terminal width so the table fills the available width instead of
// spilling into unaligned, terminal-wrapped lines for long values.
func renderTable(w io.Writer, data any) error {
	rows, err := normalizeRows(data)
	if err != nil {
		return err
	}
	rows = wrapScalarRows(rows)

	columns := deriveColumns(rows)
	if len(columns) == 0 {
		return nil
	}

	// tabwriter (padding=2 below) adds 2 columns of blank space after every
	// column's widest cell, so budget for that or the rendered table
	// exceeds terminalWidth() by 2*len(columns) despite each cell being
	// wrapped to colWidth.
	const tabwriterPadding = 2
	colWidth := terminalWidth()/len(columns) - tabwriterPadding
	if colWidth < minColumnWidth {
		colWidth = minColumnWidth
	}

	tw := tabwriter.NewWriter(w, 0, 8, tabwriterPadding, ' ', 0)

	headers := make([]string, len(columns))
	for i, c := range columns {
		headers[i] = strings.ToUpper(c)
	}
	writeWrappedRow(tw, headers, colWidth)

	for _, row := range rows {
		cells := make([]string, len(columns))
		for i, c := range columns {
			cells[i] = cellValue(row, c)
		}
		writeWrappedRow(tw, cells, colWidth)
	}

	return tw.Flush()
}

// writeWrappedRow wraps every cell to colWidth — including the header row,
// so a wide header (e.g. a long custom-columns name) can't force tabwriter
// to widen that column past the budget every other row wrapped to — and
// prints as many tabwriter lines as the tallest wrapped cell needs, padding
// shorter cells with blank lines so every column stays aligned.
func writeWrappedRow(tw *tabwriter.Writer, cells []string, colWidth int) {
	cellLines := make([][]string, len(cells))
	maxLines := 1
	for i, c := range cells {
		cellLines[i] = strings.Split(wrapCell(c, colWidth), "\n")
		if len(cellLines[i]) > maxLines {
			maxLines = len(cellLines[i])
		}
	}

	for line := 0; line < maxLines; line++ {
		out := make([]string, len(cells))
		for i := range cells {
			if line < len(cellLines[i]) {
				out[i] = strings.TrimRight(cellLines[i][line], " ")
			}
		}
		fmt.Fprintln(tw, strings.Join(out, "\t"))
	}
}

// wrapCell soft-wraps value to width, breaking on spaces where possible and
// falling back to a hard character cut for unbroken runs (e.g. base64 or
// PEM data) so no wrapped line exceeds width.
func wrapCell(value string, width int) string {
	if width < 1 || value == "" {
		return value
	}
	return lipgloss.NewStyle().Width(width).Render(value)
}
