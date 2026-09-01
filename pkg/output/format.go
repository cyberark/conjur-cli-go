package output

import (
	"fmt"
	"strings"
	"text/template"
)

type FormatType string

const (
	FormatJSON          FormatType = "json"
	FormatYAML          FormatType = "yaml"
	FormatPlain         FormatType = "plain"
	FormatTable         FormatType = "table"
	FormatCustomColumns FormatType = "custom-columns"
	FormatGoTemplate    FormatType = "go-template"
)

type ColumnSpec struct {
	Header string
	Path   string
}

type Format struct {
	Type           FormatType
	RawSpec        string
	Columns        []ColumnSpec
	ParsedTemplate *template.Template
}

func ParseFormat(spec string) (Format, error) {
	switch {
	case spec == "json":
		return Format{Type: FormatJSON, RawSpec: spec}, nil
	case spec == "yaml":
		return Format{Type: FormatYAML, RawSpec: spec}, nil
	case spec == "plain" || spec == "text":
		return Format{Type: FormatPlain, RawSpec: spec}, nil
	case spec == "table":
		return Format{Type: FormatTable, RawSpec: spec}, nil
	case strings.HasPrefix(spec, "custom-columns="):
		rest := strings.TrimPrefix(spec, "custom-columns=")
		cols, err := parseColumnsSpec(rest)
		if err != nil {
			return Format{}, err
		}
		return Format{Type: FormatCustomColumns, RawSpec: spec, Columns: cols}, nil
	case strings.HasPrefix(spec, "go-template="):
		rest := strings.TrimPrefix(spec, "go-template=")
		tmpl, err := parseGoTemplate(rest)
		if err != nil {
			return Format{}, err
		}
		return Format{Type: FormatGoTemplate, RawSpec: spec, ParsedTemplate: tmpl}, nil
	default:
		return Format{}, fmt.Errorf("unrecognized output format %q", spec)
	}
}
