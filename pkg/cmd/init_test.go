package cmd

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/cyberark/conjur-cli-go/pkg/testutils"

	"github.com/stretchr/testify/assert"
)

var initCmdTestCases = []struct {
	name string
	// NOTE: -f defaults to conjurrcInTmpDir in the args slice.
	// It can always be overwritten by appending another value to the args slice.
	args       []string
	out        string
	stdin      string
	beforeTest func(t *testing.T, conjurrcInTmpDir string)
	assert     func(t *testing.T, conjurrcInTmpDir string, stdout string, stderr string, err error)
}{
	{
		name: "help",
		args: []string{"--help"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string, stderr string, err error) {
			assert.Contains(
				t,
				stdout,
				"Use the init command to initialize the Conjur CLI with a Conjur endpoint.",
			)
		},
	},
	{
		name:  "prompts for account and URL",
		args:  []string{},
		stdin: "http://conjur\ndev\n",
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "Enter the URL of your Conjur service:")
			assert.Contains(t, stdout, "Enter your organization account name:")
			assert.Contains(t, stdout, "Wrote configuration to "+conjurrcInTmpDir)

			data, _ := ioutil.ReadFile(conjurrcInTmpDir)
			expectedConjurrc := `account: dev
appliance_url: http://conjur
plugins: []
`

			assert.Equal(t, expectedConjurrc, string(data))
		},
	},
	{
		name: "writes conjurrc",
		args: []string{"-u=https://host", "-a=test-account"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string, stderr string, err error) {
			data, _ := ioutil.ReadFile(conjurrcInTmpDir)
			expectedConjurrc := `account: test-account
appliance_url: https://host
plugins: []
`

			assert.Equal(t, expectedConjurrc, string(data))
			assert.Contains(t, stdout, "Wrote configuration to "+conjurrcInTmpDir)
		},
	},
	{
		name:  "prompts for overwrite, reject",
		args:  []string{"-u=https://host", "-a=other-test-account"},
		stdin: "N\n",
		beforeTest: func(t *testing.T, conjurrcInTmpDir string) {
			ioutil.WriteFile(conjurrcInTmpDir, []byte("something"), 0644)
		},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string, stderr string, err error) {
			// Assert that file is not overwritten
			data, _ := ioutil.ReadFile(conjurrcInTmpDir)
			assert.Equal(t, "something", string(data))

			// Assert on output
			assert.Contains(t, stdout, ".conjurrc exists. Overwrite? [y/N]")
			assert.Contains(t, stderr, "Error: Not overwriting")
		},
	},
	{
		name:  "prompts for overwrite, accept",
		args:  []string{"-u=https://host", "-a=other-test-account"},
		stdin: "y\n",
		beforeTest: func(t *testing.T, conjurrcInTmpDir string) {
			ioutil.WriteFile(conjurrcInTmpDir, []byte("something"), 0644)
		},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string, stderr string, err error) {
			// Assert that file is overwritten
			data, _ := ioutil.ReadFile(conjurrcInTmpDir)
			expectedConjurrc := `account: other-test-account
appliance_url: https://host
plugins: []
`
			assert.Equal(t, expectedConjurrc, string(data))

			// Assert on output
			assert.Contains(t, stdout, ".conjurrc exists. Overwrite? [y/N]")
			assert.Contains(t, stdout, "Wrote configuration to "+conjurrcInTmpDir)
		},
	},
	{
		name: "errors on missing conjurrc file directory",
		args: []string{"-u=https://host", "-a=test-account", "-f=/no/such/dir/file"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string, stderr string, err error) {
			assert.Contains(t, stderr, "no such file or directory")
		},
	},
}

func TestInitCmd(t *testing.T) {
	for _, tc := range initCmdTestCases {
		t.Run(tc.name, func(t *testing.T) {
			// conjurrcInTmpDir is the .conjurrc test file location, it is
			// cleaned up before every test
			conjurrcInTmpDir := t.TempDir() + "/.conjurrc"
			os.Remove(conjurrcInTmpDir)
			defer os.Remove(conjurrcInTmpDir)

			if tc.beforeTest != nil {
				tc.beforeTest(t, conjurrcInTmpDir)
			}

			cmd := NewInitCommand()

			// -f default to conjurrcInTmpDir. It can always be overwritten in each test case
			args := []string{"-f=" + conjurrcInTmpDir}
			args = append(args, tc.args...)

			stdout, stderr, err := testutils.ExecuteWithStdIn(t, cmd, tc.stdin, args...)
			tc.assert(t, conjurrcInTmpDir, stdout, stderr, err)
		})
	}

	// Other tests
	t.Run("default flags", func(t *testing.T) {
		cmd := NewInitCommand()
		cmd.ParseFlags(nil)

		f, _ := cmd.Flags().GetString("file")
		assert.Equal(t, "/root/.conjurrc", f)
	})
}
