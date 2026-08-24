package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/template"
)

func templateFuncMap() template.FuncMap {
	return template.FuncMap{
		"json": func(v any) (string, error) {
			b, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return string(b), nil
		},
		"lower": strings.ToLower,
		"upper": strings.ToUpper,
		// join accepts []string directly, but since renderGoTemplate
		// round-trips data through JSON, slices typically arrive as
		// []any (json.Unmarshal decodes JSON arrays into []interface{}).
		// Handle both so `{{ join .items "," }}` works against
		// JSON-round-tripped data as well as raw Go values.
		"join": func(items any, sep string) (string, error) {
			switch v := items.(type) {
			case []string:
				return strings.Join(v, sep), nil
			case []any:
				parts := make([]string, len(v))
				for i, item := range v {
					parts[i] = fmt.Sprint(item)
				}
				return strings.Join(parts, sep), nil
			default:
				return "", fmt.Errorf("join: unsupported type %T", items)
			}
		},
	}
}

func parseGoTemplate(spec string) (*template.Template, error) {
	tmpl, err := template.New("output").Funcs(templateFuncMap()).Parse(spec)
	if err != nil {
		return nil, fmt.Errorf("invalid go-template: %w", err)
	}
	return tmpl, nil
}

// roundTripJSON marshals then unmarshals data so templates operate on a
// generic map/slice tree, without wrapping single objects into a slice
// (unlike normalizeRows, used by custom-columns).
func roundTripJSON(data any) (any, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("go-template: marshaling data: %w", err)
	}
	var unmarshaled any
	if err := json.Unmarshal(raw, &unmarshaled); err != nil {
		return nil, fmt.Errorf("go-template: unmarshaling data: %w", err)
	}
	return unmarshaled, nil
}

func renderGoTemplate(w io.Writer, tmpl *template.Template, data any) error {
	unmarshaled, err := roundTripJSON(data)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, unmarshaled); err != nil {
		return fmt.Errorf("executing go-template: %w", err)
	}

	if !bytes.HasSuffix(buf.Bytes(), []byte("\n")) {
		buf.WriteByte('\n')
	}

	_, err = w.Write(buf.Bytes())
	return err
}
