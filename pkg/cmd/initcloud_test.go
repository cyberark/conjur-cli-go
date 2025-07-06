package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/stretchr/testify/assert"
)

var initCloudCmdTestCases = []struct {
	name string
	// NOTE: -f defaults to conjurrcInTmpDir in the args slice.
	// It can always be overwritten by appending another value to the args slice.
	args            []string
	out             string
	promptResponses []promptResponse
	// Being unable to pipe responses to prompts is a known shortcoming of Survey.
	// https://github.com/go-survey/survey/issues/394
	// This flag is used to enable Pipe-based, and not PTY-based, tests.
	pipe            bool
	beforeTest      func(t *testing.T, conjurrcInTmpDir string) func()
	assert          func(t *testing.T, conjurrcInTmpDir string, stdout string)
	jwtAuthenticate func(t *testing.T, client clients.ConjurClient) error
}{
	{
		name: "help",
		args: []string{"init", "cloud", "--help"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "writes conjurrc",
		args: []string{"init", "cloud", "-u=https://tenant.secretsmgr.cyberark.cloud", "--ca-cert=conjur-server.pem"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			data, _ := os.ReadFile(conjurrcInTmpDir)
			pwd, _ := os.Getwd()
			expectedConjurrc := `account: conjur
appliance_url: https://tenant.secretsmgr.cyberark.cloud/api
cert_file: ` + pwd + `/conjur-server.pem
authn_type: cloud
service_id: cyberark
environment: cloud
`
			assert.Equal(t, expectedConjurrc, string(data))
			assert.Contains(t, stdout, "Wrote configuration to "+conjurrcInTmpDir)
			assert.Contains(t, stdout, "Wrote certificate to")
		},
	},
	{
		name: "prompts for URL",
		args: []string{"init", "cloud", "--ca-cert=conjur-server.pem"},
		promptResponses: []promptResponse{{
			prompt:   "Enter the URL of your Conjur service:",
			response: "https://tenant.secretsmgr.cyberark.cloud",
		}},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "Wrote configuration to "+conjurrcInTmpDir)
			data, _ := os.ReadFile(conjurrcInTmpDir)
			pwd, _ := os.Getwd()
			expectedConjurrc := `account: conjur
appliance_url: https://tenant.secretsmgr.cyberark.cloud/api
cert_file: ` + pwd + `/conjur-server.pem
authn_type: cloud
service_id: cyberark
environment: cloud
`
			assert.Equal(t, expectedConjurrc, string(data))
		},
	},
	{
		name: "prompts for overwrite, reject",
		args: []string{"init", "cloud", "-u=https://tenant.secretsmgr.cyberark.cloud", "--ca-cert=conjur-server.pem"},
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
		args: []string{"init", "cloud", "-u=https://tenant.secretsmgr.cyberark.cloud", "--ca-cert=conjur-server.pem"},
		promptResponses: []promptResponse{{
			prompt:   ".conjurrc exists. Overwrite?",
			response: "y",
		}},
		pipe: true,
		beforeTest: func(t *testing.T, conjurrcInTmpDir string) func() {
			os.WriteFile(conjurrcInTmpDir, []byte("something"), 0644)
			return nil
		},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			// Assert that file is overwritten
			data, _ := os.ReadFile(conjurrcInTmpDir)
			pwd, _ := os.Getwd()
			expectedConjurrc := `account: conjur
appliance_url: https://tenant.secretsmgr.cyberark.cloud/api
cert_file: ` + pwd + `/conjur-server.pem
authn_type: cloud
service_id: cyberark
environment: cloud
`
			assert.Equal(t, expectedConjurrc, string(data))
		},
	},
	{
		name: "writes conjurrc with force netrc",
		args: []string{"init", "cloud", "-u=https://tenant.secretsmgr.cyberark.cloud", "--force-netrc", "--ca-cert=conjur-server.pem"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			data, _ := os.ReadFile(conjurrcInTmpDir)
			pwd, _ := os.Getwd()
			expectedConjurrc := `account: conjur
appliance_url: https://tenant.secretsmgr.cyberark.cloud/api
cert_file: ` + pwd + `/conjur-server.pem
authn_type: cloud
service_id: cyberark
credential_storage: file
environment: cloud
`

			assert.Equal(t, expectedConjurrc, string(data))
			assert.Contains(t, stdout, "Wrote configuration to "+conjurrcInTmpDir)
		},
	},
	{
		name: "force overwrite",
		args: []string{"init", "cloud", "-u=https://tenant.secretsmgr.cyberark.cloud", "--force", "--ca-cert=conjur-server.pem"},
		beforeTest: func(t *testing.T, conjurrcInTmpDir string) func() {
			os.WriteFile(conjurrcInTmpDir, []byte("something"), 0644)
			return nil
		},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			// Assert that file is overwritten
			data, _ := os.ReadFile(conjurrcInTmpDir)
			pwd, _ := os.Getwd()
			expectedConjurrc := `account: conjur
appliance_url: https://tenant.secretsmgr.cyberark.cloud/api
cert_file: ` + pwd + `/conjur-server.pem
authn_type: cloud
service_id: cyberark
environment: cloud
`
			assert.Equal(t, expectedConjurrc, string(data))

			// Assert on output
			assert.NotContains(t, stdout, ".conjurrc exists. Overwrite?")
			assert.Contains(t, stdout, "Wrote configuration to "+conjurrcInTmpDir)
		},
	},
	{
		name: "errors on missing conjurrc file directory",
		args: []string{"init", "cloud", "-u=https://tenant.secretsmgr.cyberark.cloud", "-f=/no/such/dir/file", "--ca-cert=conjur-server.pem"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "no such file or directory")
		},
	},
	{
		name: "writes configuration",
		args: []string{"init", "cloud", "-u=https://scauap.integration-cyberark.cloud"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "Wrote configuration to "+conjurrcInTmpDir)
		},
	},
	{
		name: "fails if can't retrieve server certificate",
		args: []string{"init", "cloud", "-u=https://nonexisting.integration-cyberark.cloud"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "Unable to retrieve and validate certificate")
			assertFetchCertFailed(t, conjurrcInTmpDir)
		},
	},
	{
		name: "fails for http urls",
		args: []string{"init", "cloud", "-u=http://example.com"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "Error: Conjur Cloud URL must use HTTPS")
			assertFetchCertFailed(t, conjurrcInTmpDir)
		},
	},
	{
		name: "fails urls without scheme",
		args: []string{"init", "cloud", "-u=invalid-url"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "Error: Conjur Cloud URL must use HTTPS")
			assertFetchCertFailed(t, conjurrcInTmpDir)
		},
	},
	{
		name: "fails for invalid urls",
		args: []string{"init", "cloud", "-u=https://invalid:url:test"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "Error: invalid Conjur Cloud URL: https://invalid:url:test/api")
			assertFetchCertFailed(t, conjurrcInTmpDir)
		},
	},
	{
		name: "allows cert specified by --ca-cert",
		args: []string{"init", "cloud", "-u=https://tenant.secretsmgr.cyberark.cloud", "--ca-cert=custom.pem"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			data, _ := os.ReadFile(conjurrcInTmpDir)
			pwd, _ := os.Getwd()
			expectedConjurrc := `account: conjur
appliance_url: https://tenant.secretsmgr.cyberark.cloud/api
cert_file: ` + pwd + `/custom.pem
authn_type: cloud
service_id: cyberark
environment: cloud
`
			assert.Equal(t, expectedConjurrc, string(data))
		},
	}, {
		name: "works for CC",
		args: []string{"init", "CC", "-u=https://tenant.secretsmgr.cyberark.cloud", "--ca-cert=conjur-server.pem"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			data, _ := os.ReadFile(conjurrcInTmpDir)
			pwd, _ := os.Getwd()
			expectedConjurrc := `account: conjur
appliance_url: https://tenant.secretsmgr.cyberark.cloud/api
cert_file: ` + pwd + `/conjur-server.pem
authn_type: cloud
service_id: cyberark
environment: cloud
`
			assert.Equal(t, expectedConjurrc, string(data))
		},
	},
}

func TestInitCloudCmd(t *testing.T) {
	for _, tc := range initCloudCmdTestCases {
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
				// "--ca-cert=" + tempDir + "/conjur-server.pem",
			}
			args = append(args, tc.args...)

			cmd := newCloudInitCmd()
			rootCmd := newRootCommand()
			initCmd := newInitCommand()
			initCmd.AddCommand(cmd)
			rootCmd.AddCommand(initCmd)
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

		cmd := newInitEnterpriseCommand(defaultInitCmdFuncs)
		rootCmd := newRootCommand()
		rootCmd.AddCommand(cmd)
		rootCmd.SetArgs([]string{"init"})
		executeCommandForTest(t, rootCmd)

		f, err := cmd.Flags().GetString("file")
		assert.NoError(t, err)
		homeDir, err := os.UserHomeDir()
		assert.NoError(t, err)
		assert.Equal(t, homeDir+"/.conjurrc", f)

		f, err = cmd.Flags().GetString("cert-file")
		assert.NoError(t, err)
		assert.Equal(t, homeDir+"/conjur-server.pem", f)
	})
}

func Test_normalizeCloudURL(t *testing.T) {
	tests := []struct {
		name         string
		applianceURL string
		want         string
	}{{
		name:         "normalize",
		applianceURL: "https://conjur",
		want:         "https://conjur/api",
	}, {
		name:         "normalize with trailing slash",
		applianceURL: "https://conjur/",
		want:         "https://conjur/api",
	}, {
		name:         "normalize with trailing slash and path",
		applianceURL: "https://conjur/api/",
		want:         "https://conjur/api",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, normalizeCloudURL(tt.applianceURL), "normalizeCloudURL(%v)", tt.applianceURL)
		})
	}
}
