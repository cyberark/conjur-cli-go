package test

import (
	"bytes"

	"testing"

	"github.com/spf13/cobra"
)

// Execute the given command and return its output.
func Execute(t *testing.T, c *cobra.Command, args ...string) (string, error) {
	t.Helper()

	buf := new(bytes.Buffer)
	c.SetOut(buf)
	c.SetErr(buf)
	c.SetArgs(append([]string{"--help=false"}, args...))

	err := c.Execute()

	return buf.String(), err
}
