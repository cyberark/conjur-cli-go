package output

import (
	"fmt"
	"io"
)

// Render dispatches on format.Type to render data via the appropriate
// renderer. It is the package's public entry point.
//
// FormatPlain has no generic renderer here: per spec, plain text shape is
// caller-supplied, so Render returns an error rather than silently no-op-ing.
func Render(w io.Writer, format Format, data any) error {
	switch format.Type {
	case FormatJSON:
		return renderJSON(w, data)
	case FormatYAML:
		return renderYAML(w, data)
	case FormatTable:
		return renderTable(w, data)
	case FormatCustomColumns:
		return renderCustomColumns(w, format.Columns, data)
	case FormatGoTemplate:
		if format.ParsedTemplate == nil {
			return fmt.Errorf("go-template format requires a parsed template")
		}
		return renderGoTemplate(w, format.ParsedTemplate, data)
	case FormatPlain:
		return fmt.Errorf("output.Render does not support FormatPlain; callers must render plain text themselves")
	default:
		return fmt.Errorf("unsupported format type %q", format.Type)
	}
}
