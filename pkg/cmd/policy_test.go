package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/spf13/cobra"

	"github.com/stretchr/testify/assert"
)

type loadPolicyTestFunc func(
	t *testing.T,
	mode conjurapi.PolicyMode,
	policyBranch string, policySrc io.Reader,
) (*conjurapi.PolicyResponse, error)

type mockPolicyClient struct {
	t          *testing.T
	loadPolicy loadPolicyTestFunc
}

func (m mockPolicyClient) LoadPolicy(
	mode conjurapi.PolicyMode,
	policyBranch string,
	policySrc io.Reader,
) (*conjurapi.PolicyResponse, error) {
	return m.loadPolicy(m.t, mode, policyBranch, policySrc)
}

type policyCmdTestCase struct {
	name string
	args []string // $TMPFILE in any of the args is substituted
	// for the temporary file created for each test
	beforeTest         func(t *testing.T, pathToTmpfile string)
	loadPolicy         loadPolicyTestFunc
	promptResponses    []promptResponse
	clientFactoryError error
	assert             func(t *testing.T, stdout string, stderr string, err error)
}

var policyCmdTestCases = []policyCmdTestCase{
	{
		name: "policy command help",
		args: []string{"policy", "--help"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
}

func sharedPolicyCmdTestCases(
	subcommand string,
	expectedMode conjurapi.PolicyMode,
) []policyCmdTestCase {
	return []policyCmdTestCase{
		{
			name: fmt.Sprintf("%s subcommand help", subcommand),
			args: []string{"policy", subcommand, "--help"},
			assert: func(t *testing.T, stdout, stderr string, err error) {
				assert.Contains(t, stdout, "HELP LONG")
			},
		},
		{
			name: fmt.Sprintf("%s subcommand policy mode", subcommand),
			args: []string{"policy", subcommand, "-b", "meow", "-f", "-"},
			loadPolicy: func(
				t *testing.T,
				mode conjurapi.PolicyMode,
				policyBranch string,
				policySrc io.Reader,
			) (*conjurapi.PolicyResponse, error) {
				// Assert on mode
				assert.Equal(t, expectedMode, mode)

				return nil, nil
			},
		},
		{
			name: fmt.Sprintf("%s subcommand with good response", subcommand),
			args: []string{"policy", subcommand, "-b", "meow", "-f", "-"},
			loadPolicy: func(
				t *testing.T,
				mode conjurapi.PolicyMode,
				policyBranch string,
				policySrc io.Reader,
			) (*conjurapi.PolicyResponse, error) {
				return &conjurapi.PolicyResponse{
					CreatedRoles: map[string]conjurapi.CreatedRole{
						"a role": {
							ID:     "a role id",
							APIKey: "a role api key",
						},
					},
					Version: 1234,
				}, nil
			},
			assert: func(t *testing.T, stdout, stderr string, err error) {
				assert.Contains(t, stdout, "created_roles")
				assert.Contains(t, stdout, "version")
				assert.Contains(t, stderr, "Loaded policy 'meow'")
			},
		},
		{
			name: fmt.Sprintf("%s subcommand from stdin", subcommand),
			promptResponses: []promptResponse{
				{
					prompt:   "", // An empty prompt means to immediately write to stdin
					response: "policy file content\n",
				},
			},
			args: []string{"policy", subcommand, "-b", "meow", "-f", "-"},
			loadPolicy: func(
				t *testing.T,
				mode conjurapi.PolicyMode,
				policyBranch string,
				policySrc io.Reader,
			) (*conjurapi.PolicyResponse, error) {
				policyContents, err := io.ReadAll(policySrc)
				assert.NoError(t, err)
				assert.Equal(t, "policy file content\n", string(policyContents))

				return nil, nil
			},
			assert: func(t *testing.T, stdout, stderr string, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: fmt.Sprintf("%s subcommand from file", subcommand),
			args: []string{"policy", subcommand, "-b", "meow", "-f", "$TMPFILE"},
			beforeTest: func(t *testing.T, pathToTmpfile string) {
				err := os.WriteFile(pathToTmpfile, []byte("policy file content"), 0644)
				assert.NoError(t, err)
			},
			loadPolicy: func(
				t *testing.T,
				mode conjurapi.PolicyMode,
				policyBranch string,
				policySrc io.Reader,
			) (*conjurapi.PolicyResponse, error) {
				policyContents, err := io.ReadAll(policySrc)
				assert.NoError(t, err)
				assert.Equal(t, "policy file content", string(policyContents))

				return nil, nil
			},
			assert: func(t *testing.T, stdout, stderr string, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: fmt.Sprintf("%s subcommand response error", subcommand),
			args: []string{"policy", subcommand, "-b", "meow", "-f", "-"},
			loadPolicy: func(
				t *testing.T,
				mode conjurapi.PolicyMode,
				policyBranch string,
				policySrc io.Reader,
			) (*conjurapi.PolicyResponse, error) {
				return nil, fmt.Errorf("%s", "some error")
			},
			assert: func(t *testing.T, stdout, stderr string, err error) {
				assert.Contains(t, stderr, "Error: some error")
			},
		},
		{
			name:               fmt.Sprintf("%s subcommand client factory error", subcommand),
			args:               []string{"policy", subcommand, "-b", "meow", "-f", "-"},
			clientFactoryError: fmt.Errorf("client factory error"),
			assert: func(t *testing.T, stdout, stderr string, err error) {
				assert.Contains(t, stderr, "Error: client factory error")
			},
		},
		{
			name: fmt.Sprintf("%s subcommand missing filepath", subcommand),
			args: []string{"policy", subcommand, "-b", "meow"},
			assert: func(t *testing.T, stdout, stderr string, err error) {
				assert.Contains(t, stderr, "Error: required flag(s) \"filepath\" not set\n")
			},
		},
		{
			name: fmt.Sprintf("%s subcommand missing branch", subcommand),
			args: []string{"policy", subcommand, "-f", "-"},
			assert: func(t *testing.T, stdout, stderr string, err error) {
				assert.Contains(t, stderr, "Error: required flag(s) \"branch\" not set\n")
			},
		},
	}
}

func TestPolicyCmd(t *testing.T) {
	t.Parallel()

	var allTests []policyCmdTestCase
	for _, cases := range [][]policyCmdTestCase{
		policyCmdTestCases,
		sharedPolicyCmdTestCases(
			"load",
			conjurapi.PolicyModePost,
		),
		sharedPolicyCmdTestCases(
			"append",
			conjurapi.PolicyModePatch,
		),
		sharedPolicyCmdTestCases(
			"replace",
			conjurapi.PolicyModePut,
		),
	} {
		allTests = append(allTests, cases...)
	}

	for _, tc := range allTests {
		t.Run(tc.name, func(t *testing.T) {
			pathToTmpfile := t.TempDir() + "/file"
			os.Remove(pathToTmpfile)
			defer os.Remove(pathToTmpfile)

			if tc.beforeTest != nil {
				tc.beforeTest(t, pathToTmpfile)
			}

			mockClient := mockPolicyClient{t: t, loadPolicy: tc.loadPolicy}

			cmd := newPolicyCommand(
				func(cmd *cobra.Command) (policyClient, error) {
					return mockClient, tc.clientFactoryError
				},
			)

			// $TMPFILE points to tempfile created for each test run
			for i, v := range tc.args {
				tc.args[i] = strings.Replace(v, "$TMPFILE", pathToTmpfile, 1)
			}

			if tc.promptResponses != nil {
				// Use the prompt responses helper to simulate user input
				stdout, err := executeCommandForTestWithPromptResponses(
					t, cmd, tc.promptResponses,
				)
				if tc.assert != nil {
					tc.assert(t, stdout, "", err)
				}
			} else {
				stdout, stderr, err := executeCommandForTest(
					t, cmd, tc.args...,
				)
				if tc.assert != nil {
					tc.assert(t, stdout, stderr, err)
				}
			}
		})
	}
}
