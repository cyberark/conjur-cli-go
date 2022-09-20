package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func execute(t *testing.T, c *cobra.Command, args ...string) (string, error) {
	t.Helper()

	buf := new(bytes.Buffer)
	c.SetOut(buf)
	c.SetErr(buf)
	c.SetArgs(args)

	err := c.Execute()

	return buf.String(), err
}

func TestInitCmd(t *testing.T) {
	conjurrcInTmpDir := t.TempDir() + "/.conjurrc"
	defer os.Remove(conjurrcInTmpDir)

	tt := []struct {
		args []string
		out  string
	}{
		{
			args: []string{"--help"},
			out:  "Initialize the Conjur configuration",
		},
		{
			args: []string{"--helpx"},
			out:  "unknown flag: --helpx",
		},
		{
			args: []string{"--help=false"},
			out:  "Enter the URL of your Conjur service",
		},
		{
			args: []string{"--help=false", "-u=https://host"},
			out:  "Enter your organization account name",
		},
		{
			args: []string{"--help=false", "-u=https://host", "-a=test-account", "-f=" + conjurrcInTmpDir},
			out:  "Wrote configuration to " + conjurrcInTmpDir,
		},
		{
			args: []string{"--help=false", "-u=https://host", "-a=test-account", "-f=/no/such/dir/file"},
			out:  "no such file or directory",
		},
	}

	cmd := NewInitCommand()

	for _, tc := range tt {
		out, _ := execute(t, cmd, tc.args...)

		assert.Contains(t, out, tc.out)
	}
}
