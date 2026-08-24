// pkg/output/yaml.go
package output

import (
	"io"

	"gopkg.in/yaml.v3"
)

func renderYAML(w io.Writer, data any) error {
	enc := yaml.NewEncoder(w)
	if err := enc.Encode(data); err != nil {
		enc.Close()
		return err
	}
	return enc.Close()
}
