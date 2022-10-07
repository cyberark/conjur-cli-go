package test

import (
	"bytes"

	"github.com/spf13/cobra"
	"testing"
)

// Execute the given command and return its output.
func Execute(t *testing.T, c *cobra.Command, args ...string) (string, error) {
	t.Helper()

	buf := new(bytes.Buffer)
	c.SetOut(buf)
	c.SetErr(buf)
	c.SetArgs(args)

	err := c.Execute()

	return buf.String(), err
}
