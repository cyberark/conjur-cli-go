package cmd

import (
	"fmt"
	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
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
	pipe            bool
	beforeTest      func(t *testing.T, conjurrcInTmpDir string) func()
	assert          func(t *testing.T, conjurrcInTmpDir string, stdout string)
	jwtAuthenticate func(t *testing.T, client clients.ConjurClient) error
}{
	{
		name: "help cloud",
		args: []string{"init", "--env", "cloud", "--help"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	}, {
		name: "help enterprise",
		args: []string{"init", "--env", "CE", "--help"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	}, {
		name: "help open-source",
		args: []string{"init", "--env", "open-source", "--help"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	}, {
		name: "help prompts for environment",
		args: []string{"init", "--help"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "HELP LONG")
		},
		promptResponses: []promptResponse{{
			prompt:   "? Select the environment you want to use:  [Use arrows to move, type to filter, ? for more help]",
			response: "cloud\n",
		}},
		pipe: true,
	}, {
		name: "env flag redirects help to subcommand",
		args: []string{"init", "--env", "cloud", "--help"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "HELP LONG")
			assert.Contains(t, stdout, "conjur init saas")
		},
	}, {
		name: "env flag redirects command to subcommand",
		args: []string{"init", "--env", "cloud", "-u=http://host"},
		assert: func(t *testing.T, conjurrcInTmpDir string, stdout string) {
			assert.Contains(t, stdout, "Error: Secrets Manager SaaS URL must use HTTPS")
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
			mockClient := mockInitClient{t: t, jwtAuthenticate: tc.jwtAuthenticate}

			initEnt := newInitEnterpriseCommand(initCmdFuncs{
				JWTAuthenticate: mockClient.JWTAuthenticate,
			})
			rootCmd := newRootCommand()
			initCmd := newInitCommand()
			initCmd.AddCommand(newCloudInitCmd())
			initCmd.AddCommand(initEnt)
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

func Test_getEnv(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{{
		"cloud with space",
		[]string{"init", "--env", "cloud"},
		"cloud",
	}, {
		"oss with equals",
		[]string{"init", "--env=oss"},
		"oss",
	}, {
		"missing env value",
		[]string{"init", "--env"},
		"",
	}, {
		"missing env flag",
		[]string{"init", "--file=/tmp/file"},
		"",
	}, {
		name: "empty",
		args: []string{},
		want: "",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, getEnv(tt.args), "getEnv(%v)", tt.args)
		})
	}
}
