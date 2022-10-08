package testutils

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
	// The help flag is somehow defaulting to true. The fix is to prepend the args with --help=false.
	c.SetArgs(append([]string{"--help=false"}, args...))

	err := c.Execute()

	return buf.String(), err
}
