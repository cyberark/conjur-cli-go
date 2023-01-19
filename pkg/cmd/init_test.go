package cmd

import (
	"io"
	"os"
	"path/filepath"
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
		args: []string{"init", "-i"},
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
		args: []string{"init", "-u=http://host", "-a=test-account", "-i"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string, stderr string, err error) {
			data, _ := os.ReadFile(conjurrcInTmpDir)
			expectedConjurrc := `account: test-account
appliance_url: http://host
`

			assert.Equal(t, expectedConjurrc, string(data))
			assert.Contains(t, stdout, "Wrote configuration to "+conjurrcInTmpDir)
			// Shouldn't write certificate for HTTP url
			assert.NotContains(t, stdout, "Wrote certificate to")
		},
	},
	{
		name: "writes conjurrc for ldap",
		args: []string{"init", "-u=http://host", "-a=test-account", "-t=ldap", "--service-id=test", "-i"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string, stderr string, err error) {
			data, _ := os.ReadFile(conjurrcInTmpDir)
			expectedConjurrc := `account: test-account
appliance_url: http://host
authn_type: ldap
service_id: test
`

			assert.Equal(t, expectedConjurrc, string(data))
			assert.Contains(t, stdout, "Wrote configuration to "+conjurrcInTmpDir)
		},
	},
	{
		name: "prompts for overwrite, reject",
		args: []string{"init", "-u=http://host", "-a=other-test-account", "-i"},
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
		args: []string{"init", "-u=http://host", "-a=other-test-account", "-i"},
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
appliance_url: http://host
`
			assert.Equal(t, expectedConjurrc, string(data))

			// Assert on output
			assert.Contains(t, stdout, ".conjurrc exists. Overwrite? [y/N]")
			assert.Contains(t, stdout, "Wrote configuration to "+conjurrcInTmpDir)
		},
	},
	{
		name: "force overwrite",
		args: []string{"init", "-u=http://host", "-a=yet-another-test-account", "--force", "-i"},
		beforeTest: func(t *testing.T, conjurrcInTmpDir string) {
			os.WriteFile(conjurrcInTmpDir, []byte("something"), 0644)
		},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string, stderr string, err error) {
			// Assert that file is overwritten
			data, _ := os.ReadFile(conjurrcInTmpDir)
			expectedConjurrc := `account: yet-another-test-account
appliance_url: http://host
`
			assert.Equal(t, expectedConjurrc, string(data))

			// Assert on output
			assert.NotContains(t, stdout, ".conjurrc exists. Overwrite? [y/N]")
			assert.Contains(t, stdout, "Wrote configuration to "+conjurrcInTmpDir)
		},
	},
	{
		name: "errors on missing conjurrc file directory",
		args: []string{"init", "-u=http://host", "-a=test-account", "-f=/no/such/dir/file", "-i"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string, stderr string, err error) {
			assert.Contains(t, stderr, "no such file or directory")
		},
	},
	{
		name: "writes certificate",
		args: []string{"init", "-u=https://example.com", "-a=test-account"},
		promptResponses: []promptResponse{
			{
				prompt:   "Trust this certificate? [y/N]",
				response: "y",
			},
		},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string, stderr string, err error) {
			assert.NoError(t, err)
			assert.Equal(t, "", stderr)
			assertCertWritten(t, conjurrcInTmpDir, stdout)
		},
	},
	{
		name: "prompts to trust certificate, reject",
		args: []string{"init", "-u=https://example.com", "-a=test-account"},
		promptResponses: []promptResponse{
			{
				prompt:   "Trust this certificate? [y/N]",
				response: "n",
			},
		},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string, stderr string, err error) {
			assert.Contains(t, stderr, "You decided not to trust the certificate")
			assertFetchCertFailed(t, conjurrcInTmpDir)
		},
	},
	{
		name: "fails if can't retrieve server certificate",
		args: []string{"init", "-u=https://nohost.example.com", "-a=test-account"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string, stderr string, err error) {
			assert.Contains(t, stderr, "Unable to retrieve certificate")
			assertFetchCertFailed(t, conjurrcInTmpDir)
		},
	},
	{
		name: "fails for self-signed certificate",
		args: []string{"init", "-u=https://self-signed.badssl.com", "-a=test-account"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string, stderr string, err error) {
			assert.Contains(t, stderr, "Unable to retrieve certificate")
			assertFetchCertFailed(t, conjurrcInTmpDir)
		},
	},
	{
		name: "allows self-signed certificate when specified",
		args: []string{"init", "-u=https://self-signed.badssl.com", "-a=test-account", "--self-signed"},
		promptResponses: []promptResponse{
			{
				prompt:   "Trust this certificate? [y/N]",
				response: "y",
			},
		},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string, stderr string, err error) {
			assert.NoError(t, err)
			assert.Equal(t, "Warning: Using self-signed certificates is not recommended and could lead to exposure of sensitive data\n", stderr)
			assertCertWritten(t, conjurrcInTmpDir, stdout)
		},
	},
	{
		name: "fails for http urls",
		args: []string{"init", "-u=http://example.com", "-a=test-account"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string, stderr string, err error) {
			assert.Contains(t, stderr, "Cannot fetch certificate from non-HTTPS URL")
			assertFetchCertFailed(t, conjurrcInTmpDir)
		},
	},
	{
		name: "allows http urls when specified",
		args: []string{"init", "-u=http://example.com", "-a=test-account", "--insecure"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string, stderr string, err error) {
			assert.NoError(t, err)
			assert.Contains(t, stderr, "Warning: Running the command with '--insecure' makes your system vulnerable to security attacks")
			assert.Contains(t, stderr, "If you prefer to communicate with the server securely you must reinitialize the client in secure mode.")
		},
	},
	{
		name: "fails if both --insecure and --self-signed are specified",
		args: []string{"init", "-u=http://example.com", "-a=test-account", "--insecure", "--self-signed"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string, stderr string, err error) {
			assert.Contains(t, stderr, "Cannot specify both --insecure and --self-signed")
			assertFetchCertFailed(t, conjurrcInTmpDir)
		},
	},
}

func TestInitCmd(t *testing.T) {
	t.Parallel()

	for _, tc := range initCmdTestCases {
		t.Run(tc.name, func(t *testing.T) {
			// conjurrcInTmpDir is the .conjurrc test file location, it is
			// cleaned up before every test
			tempDir := t.TempDir()
			conjurrcInTmpDir := tempDir + "/.conjurrc"

			if tc.beforeTest != nil {
				tc.beforeTest(t, conjurrcInTmpDir)
			}

			cmd := newInitCommand()

			// -f default to conjurrcInTmpDir. It can always be overwritten in each test case
			args := []string{
				"-f=" + tempDir + "/.conjurrc",
				"--cert-file=" + tempDir + "/conjur-server.pem",
			}
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

		f, err = cmd.Flags().GetString("cert-file")
		assert.NoError(t, err)
		assert.Equal(t, "/root/conjur-server.pem", f)
	})

	t.Run("version flag", func(t *testing.T) {
		rootCmd := newRootCommand()
		stdout, stderr, err := executeCommandForTest(t, rootCmd, "--version")

		assert.NoError(t, err)
		assert.Equal(t, "", stderr)
		assert.Equal(t, "Conjur CLI version unset-unset\n", stdout)
	})
}

func assertFetchCertFailed(t *testing.T, conjurrcInTmpDir string) {
	// Assert that conjurrc and certificate were not written
	expectedCertPath := filepath.Dir(conjurrcInTmpDir) + "/conjur-server.pem"
	_, err := os.Stat(conjurrcInTmpDir)
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(expectedCertPath)
	assert.True(t, os.IsNotExist(err))
}

func assertCertWritten(t *testing.T, conjurrcInTmpDir string, stdout string) {
	expectedCertPath := filepath.Dir(conjurrcInTmpDir) + "/conjur-server.pem"

	// Assert that the certificate path is written to conjurrc
	data, _ := os.ReadFile(conjurrcInTmpDir)
	assert.Contains(t, string(data), "cert_file: "+expectedCertPath)

	// Assert that certificate is written
	assert.Contains(t, stdout, "Wrote certificate to "+expectedCertPath+"\n")
	data, _ = os.ReadFile(expectedCertPath)
	assert.Contains(t, string(data), "-----BEGIN CERTIFICATE-----")
}
