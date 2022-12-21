package cmd

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var initCmdTestCases = []struct {
	name string
	// NOTE: -f defaults to conjurrcInTmpDir in the args slice.
	// It can always be overwritten by appending another value to the args slice.
	args            []string
	out             string
	promptResponses []promptResponse
	beforeTest      func(t *testing.T, conjurrcInTmpDir string)
	assert          func(t *testing.T, conjurrcInTmpDir string, stdout string, stderr string, err error)
}{
	{
		name: "help",
		args: []string{"init", "--help"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string, stderr string, err error) {
			assert.Contains(
				t,
				stdout,
				"HELP LONG",
			)
		},
	},
	{
		name: "prompts for account and URL",
		args: []string{"init"},
		promptResponses: []promptResponse{
			{
				prompt:   "Enter the URL of your Conjur service:",
				response: "http://conjur",
			},
			{
				prompt:   "Enter your organization account name:",
				response: "dev",
			},
		},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "Enter the URL of your Conjur service:")
			assert.Contains(t, stdout, "Enter your organization account name:")
			assert.Contains(t, stdout, "Wrote configuration to "+conjurrcInTmpDir)

			data, _ := os.ReadFile(conjurrcInTmpDir)
			expectedConjurrc := `account: dev
appliance_url: http://conjur
`

			assert.Equal(t, expectedConjurrc, string(data))
		},
	},
	{
		name: "writes conjurrc",
		args: []string{"init", "-u=https://host", "-a=test-account"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string, stderr string, err error) {
			data, _ := os.ReadFile(conjurrcInTmpDir)
			expectedConjurrc := `account: test-account
appliance_url: https://host
`

			assert.Equal(t, expectedConjurrc, string(data))
			assert.Contains(t, stdout, "Wrote configuration to "+conjurrcInTmpDir)
		},
	},
	{
		name: "writes conjurrc for ldap",
		args: []string{"init", "-u=https://host", "-a=test-account", "-t=ldap", "--service-id=test"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string, stderr string, err error) {
			data, _ := os.ReadFile(conjurrcInTmpDir)
			expectedConjurrc := `account: test-account
appliance_url: https://host
authn_type: ldap
service_id: test
`

			assert.Equal(t, expectedConjurrc, string(data))
			assert.Contains(t, stdout, "Wrote configuration to "+conjurrcInTmpDir)
		},
	},
	{
		name: "prompts for overwrite, reject",
		args: []string{"init", "-u=https://host", "-a=other-test-account"},
		promptResponses: []promptResponse{
			{
				prompt:   ".conjurrc exists. Overwrite? [y/N]",
				response: "N",
			},
		},
		beforeTest: func(t *testing.T, conjurrcInTmpDir string) {
			os.WriteFile(conjurrcInTmpDir, []byte("something"), 0644)
		},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string, stderr string, err error) {
			// Assert that file is not overwritten
			data, _ := os.ReadFile(conjurrcInTmpDir)
			assert.Equal(t, "something", string(data))

			// Assert on output
			assert.Contains(t, stdout, ".conjurrc exists. Overwrite? [y/N]")
			assert.Contains(t, stderr, "Error: Not overwriting")
		},
	},
	{
		name: "prompts for overwrite, accept",
		args: []string{"init", "-u=https://host", "-a=other-test-account"},
		promptResponses: []promptResponse{
			{
				prompt:   "",
				response: "y",
			},
		},
		beforeTest: func(t *testing.T, conjurrcInTmpDir string) {
			os.WriteFile(conjurrcInTmpDir, []byte("something"), 0644)
		},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string, stderr string, err error) {
			// Assert that file is overwritten
			data, _ := os.ReadFile(conjurrcInTmpDir)
			expectedConjurrc := `account: other-test-account
appliance_url: https://host
`
			assert.Equal(t, expectedConjurrc, string(data))

			// Assert on output
			assert.Contains(t, stdout, ".conjurrc exists. Overwrite? [y/N]")
			assert.Contains(t, stdout, "Wrote configuration to "+conjurrcInTmpDir)
		},
	},
	{
		name: "force overwrite",
		args: []string{"init", "-u=https://host", "-a=yet-another-test-account", "--force"},
		beforeTest: func(t *testing.T, conjurrcInTmpDir string) {
			os.WriteFile(conjurrcInTmpDir, []byte("something"), 0644)
		},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string, stderr string, err error) {
			// Assert that file is overwritten
			data, _ := os.ReadFile(conjurrcInTmpDir)
			expectedConjurrc := `account: yet-another-test-account
appliance_url: https://host
`
			assert.Equal(t, expectedConjurrc, string(data))

			// Assert on output
			assert.NotContains(t, stdout, ".conjurrc exists. Overwrite? [y/N]")
			assert.Contains(t, stdout, "Wrote configuration to "+conjurrcInTmpDir)
		},
	},
	{
		name: "errors on missing conjurrc file directory",
		args: []string{"init", "-u=https://host", "-a=test-account", "-f=/no/such/dir/file"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string, stderr string, err error) {
			assert.Contains(t, stderr, "no such file or directory")
		},
	},
}

func TestInitCmd(t *testing.T) {
	t.Parallel()

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

			cmd := newInitCommand()

			// -f default to conjurrcInTmpDir. It can always be overwritten in each test case
			args := []string{"-f=" + conjurrcInTmpDir}
			args = append(args, tc.args...)

			stdout, stderr, err := executeCommandForTestWithPromptResponses(
				t, cmd, tc.promptResponses, args...,
			)
			tc.assert(t, conjurrcInTmpDir, stdout, stderr, err)
		})
	}

	// Other tests
	t.Run("default flags", func(t *testing.T) {
		cmd := newInitCommand()

		rootCmd := newRootCommand()
		rootCmd.AddCommand(cmd)
		rootCmd.SetOut(io.Discard)
		rootCmd.SetErr(io.Discard)
		rootCmd.SetArgs([]string{"init"})
		rootCmd.Execute()

		f, err := cmd.Flags().GetString("file")
		assert.NoError(t, err)
		assert.Equal(t, "/root/.conjurrc", f)
	})
}
