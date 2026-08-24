// pkg/output/yaml_test.go
package output

import (
	"bytes"
	"testing"
)

func TestRenderYAML(t *testing.T) {
	data := map[string]any{"name": "foo"}
	var buf bytes.Buffer
	if err := renderYAML(&buf, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "name: foo\n"
	if buf.String() != want {
		t.Fatalf("got %q, want %q", buf.String(), want)
	}
}
