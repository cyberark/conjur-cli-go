package cmd

import (
	"os"
	"testing"

	"github.com/cyberark/conjur-cli-go/pkg/test"

	"github.com/stretchr/testify/assert"
)

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
		out, _ := test.Execute(t, cmd, tc.args...)

		assert.Contains(t, out, tc.out)
	}
}
