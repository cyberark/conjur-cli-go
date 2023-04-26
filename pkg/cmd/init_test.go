package cmd

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
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
	// Being unable to pipe responses to prompts is a known shortcoming of Survey.
	// https://github.com/go-survey/survey/issues/394
	// This flag is used to enable Pipe-based, and not PTY-based, tests.
	pipe       bool
	beforeTest func(t *testing.T, conjurrcInTmpDir string) func()
	assert     func(t *testing.T, conjurrcInTmpDir string, stdout string)
}{
	{
		name: "help",
		args: []string{"init", "--help"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "writes conjurrc",
		args: []string{"init", "-u=http://host", "-a=test-account", "-i"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
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
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
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
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "Wrote configuration to "+conjurrcInTmpDir)

			data, _ := os.ReadFile(conjurrcInTmpDir)
			expectedConjurrc := `account: dev
appliance_url: http://conjur
`

			assert.Equal(t, expectedConjurrc, string(data))
		},
	},
	{
		name: "prompts for overwrite, reject",
		args: []string{"init", "-u=http://host", "-a=other-test-account", "-i"},
		promptResponses: []promptResponse{
			{
				prompt:   ".conjurrc exists. Overwrite?",
				response: "N",
			},
		},
		pipe: true,
		beforeTest: func(t *testing.T, conjurrcInTmpDir string) func() {
			os.WriteFile(conjurrcInTmpDir, []byte("something"), 0644)
			return nil
		},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			// Assert that file is not overwritten
			data, _ := os.ReadFile(conjurrcInTmpDir)
			assert.Equal(t, "something", string(data))
		},
	},
	{
		name: "prompts for overwrite, accept",
		args: []string{"init", "-u=http://host", "-a=other-test-account", "-i"},
		promptResponses: []promptResponse{
			{
				prompt:   ".conjurrc exists. Overwrite?",
				response: "y",
			},
		},
		pipe: true,
		beforeTest: func(t *testing.T, conjurrcInTmpDir string) func() {
			os.WriteFile(conjurrcInTmpDir, []byte("something"), 0644)
			return nil
		},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			// Assert that file is overwritten
			data, _ := os.ReadFile(conjurrcInTmpDir)
			expectedConjurrc := `account: other-test-account
appliance_url: http://host
`
			assert.Equal(t, expectedConjurrc, string(data))
		},
	},
	{
		name: "writes conjurrc with force netrc",
		args: []string{"init", "-u=http://host", "-a=test-account", "--force-netrc", "-i"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			data, _ := os.ReadFile(conjurrcInTmpDir)
			expectedConjurrc := `account: test-account
appliance_url: http://host
credential_storage: file
`

			assert.Equal(t, expectedConjurrc, string(data))
			assert.Contains(t, stdout, "Wrote configuration to "+conjurrcInTmpDir)
		},
	},
	{
		name: "force overwrite",
		args: []string{"init", "-u=http://host", "-a=yet-another-test-account", "--force", "-i"},
		beforeTest: func(t *testing.T, conjurrcInTmpDir string) func() {
			os.WriteFile(conjurrcInTmpDir, []byte("something"), 0644)
			return nil
		},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			// Assert that file is overwritten
			data, _ := os.ReadFile(conjurrcInTmpDir)
			expectedConjurrc := `account: yet-another-test-account
appliance_url: http://host
`
			assert.Equal(t, expectedConjurrc, string(data))

			// Assert on output
			assert.NotContains(t, stdout, ".conjurrc exists. Overwrite?")
			assert.Contains(t, stdout, "Wrote configuration to "+conjurrcInTmpDir)
		},
	},
	{
		name: "errors on missing conjurrc file directory",
		args: []string{"init", "-u=http://host", "-a=test-account", "-f=/no/such/dir/file", "-i"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "no such file or directory")
		},
	},
	{
		name: "writes certificate",
		args: []string{"init", "-u=https://example.com", "-a=test-account"},
		promptResponses: []promptResponse{
			{
				prompt:   "Trust this certificate?",
				response: "y",
			},
		},
		pipe: true,
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assertCertWritten(t, conjurrcInTmpDir, stdout)
		},
	},
	{
		name: "prompts to trust certificate, reject",
		args: []string{"init", "-u=https://example.com", "-a=test-account"},
		promptResponses: []promptResponse{
			{
				prompt:   "Trust this certificate?",
				response: "N",
			},
		},
		pipe: true,
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			fmt.Println(stdout)
			assert.Contains(t, stdout, "You decided not to trust the certificate")
			assertFetchCertFailed(t, conjurrcInTmpDir)
		},
	},
	{
		name: "fails if can't retrieve server certificate",
		args: []string{"init", "-u=https://nohost.example.com", "-a=test-account"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "Unable to retrieve and validate certificate")
			assertFetchCertFailed(t, conjurrcInTmpDir)
		},
	},
	{
		name: "fails for self-signed certificate",
		args: []string{"init", "-u=https://localhost:8080", "-a=test-account"},
		beforeTest: func(t *testing.T, conjurrcInTmpDir string) func() {
			return startSelfSignedServer(t, 8080)
		},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "Unable to retrieve and validate certificate")
			assert.Contains(t, stdout, "x509")
			assert.Contains(t, stdout, "If you're attempting to use a self-signed certificate, re-run the init command with the `--self-signed` flag")
			assertFetchCertFailed(t, conjurrcInTmpDir)
		},
	},
	{
		name: "succeeds for self-signed certificate with --self-signed flag",
		args: []string{"init", "-u=https://localhost:8080", "-a=test-account", "--self-signed"},
		promptResponses: []promptResponse{
			{
				prompt:   "Trust this certificate?",
				response: "y",
			},
		},
		pipe: true,
		beforeTest: func(t *testing.T, conjurrcInTmpDir string) func() {
			return startSelfSignedServer(t, 8080)
		},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "Warning: Using self-signed certificates is not recommended and could lead to exposure of sensitive data")
			assertCertWritten(t, conjurrcInTmpDir, stdout)
		},
	},
	{
		name: "fails for http urls",
		args: []string{"init", "-u=http://example.com", "-a=test-account"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "Cannot fetch certificate from non-HTTPS URL")
			assertFetchCertFailed(t, conjurrcInTmpDir)
		},
	},
	{
		name: "fails urls without scheme",
		args: []string{"init", "-u=invalid-url", "-a=test-account"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "Error: Cannot fetch certificate from non-HTTPS URL")
			assertFetchCertFailed(t, conjurrcInTmpDir)
		},
	},
	{
		name: "fails for invalid urls",
		args: []string{"init", "-u=https://invalid:url:test", "-a=test-account"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "Error: parse")
			assertFetchCertFailed(t, conjurrcInTmpDir)
		},
	},
	{
		name: "allows http urls when specified",
		args: []string{"init", "-u=http://example.com", "-a=test-account", "--insecure"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "Warning: Running the command with '--insecure' makes your system vulnerable to security attacks")
			assert.Contains(t, stdout, "If you prefer to communicate with the server securely you must reinitialize the client in secure mode.")
		},
	},
	{
		name: "fails if both --insecure and --self-signed are specified",
		args: []string{"init", "-u=http://example.com", "-a=test-account", "--insecure", "--self-signed"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "Cannot specify both --insecure and --self-signed")
			assertFetchCertFailed(t, conjurrcInTmpDir)
		},
	},
	{
		name: "fails if --insecure and --ca-cert are specified",
		args: []string{"init", "-u=http://example.com", "-a=test-account", "--insecure", "--ca-cert=cert.pem"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "Cannot specify --ca-cert when using --insecure or --self-signed")
			assertFetchCertFailed(t, conjurrcInTmpDir)
		},
	},
	{
		name: "fails if --self-signed and --ca-cert are specified",
		args: []string{"init", "-u=http://example.com", "-a=test-account", "--self-signed", "--ca-cert=cert.pem"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "Cannot specify --ca-cert when using --insecure or --self-signed")
			assertFetchCertFailed(t, conjurrcInTmpDir)
		},
	},
	{
		name: "allows cert specified by --ca-cert",
		args: []string{"init", "-u=https://example.com", "-a=test-account", "--ca-cert=custom-cert.pem"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			data, _ := os.ReadFile(conjurrcInTmpDir)
			pwd, _ := os.Getwd()
			expectedConjurrc := `account: test-account
appliance_url: https://example.com
cert_file: ` + pwd + `/custom-cert.pem
`
			assert.Equal(t, expectedConjurrc, string(data))
		},
	},
}

func TestInitCmd(t *testing.T) {
	for _, tc := range initCmdTestCases {
		t.Run(tc.name, func(t *testing.T) {
			// conjurrcInTmpDir is the .conjurrc test file location, it is
			// cleaned up before every test
			tempDir := t.TempDir()
			conjurrcInTmpDir := tempDir + "/.conjurrc"

			if tc.beforeTest != nil {
				cleanup := tc.beforeTest(t, conjurrcInTmpDir)
				if cleanup != nil {
					defer cleanup()
				}
			}

			// --file default to conjurrcInTmpDir. It can always be overwritten in each test case
			args := []string{
				"--file=" + tempDir + "/.conjurrc",
				"--cert-file=" + tempDir + "/conjur-server.pem",
			}
			args = append(args, tc.args...)

			// Create command tree for init
			cmd := newInitCommand()
			rootCmd := newRootCommand()
			rootCmd.AddCommand(cmd)
			rootCmd.SetArgs(args)

			var out string
			if tc.pipe {
				var content string
				for _, pr := range tc.promptResponses {
					content = fmt.Sprintf("%s%s\n", content, pr.response)
				}
				out, _ = executeCommandForTestWithPipeResponses(t, rootCmd, content)
			} else {
				out, _ = executeCommandForTestWithPromptResponses(t, rootCmd, tc.promptResponses)
			}

			if tc.assert != nil {
				tc.assert(t, conjurrcInTmpDir, out)
			}
		},
		)
	}

	// Other tests
	t.Run("default flags", func(t *testing.T) {
		cmd := newInitCommand()

		rootCmd := newRootCommand()
		rootCmd.AddCommand(cmd)
		rootCmd.SetArgs([]string{"init"})
		executeCommandForTest(t, rootCmd)

		f, err := cmd.Flags().GetString("file")
		assert.NoError(t, err)
		assert.Equal(t, "/root/.conjurrc", f)

		f, err = cmd.Flags().GetString("cert-file")
		assert.NoError(t, err)
		assert.Equal(t, "/root/conjur-server.pem", f)
	})

	t.Run("version flag", func(t *testing.T) {
		rootCmd := newRootCommand()
		stdout, _, err := executeCommandForTest(t, rootCmd, "--version")

		assert.NoError(t, err)
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
	assert.Contains(t, stdout, "Wrote certificate to "+expectedCertPath)
	data, _ = os.ReadFile(expectedCertPath)
	assert.Contains(t, string(data), "-----BEGIN CERTIFICATE-----")
}

func startSelfSignedServer(t *testing.T, port int) func() {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		assert.NoError(t, err, "unabled to start test server")
	}

	server.Listener = l
	server.StartTLS()

	return func() { server.Close() }
}
