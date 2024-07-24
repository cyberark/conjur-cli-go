package cmd

import (
	"fmt"
	"github.com/cyberark/conjur-cli-go/pkg/utils"
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

type dryRunPolicyTestFunc func(
	t *testing.T,
	mode conjurapi.PolicyMode,
	policyBranch string, policySrc io.Reader,
) (*conjurapi.DryRunPolicyResponse, error)

type fetchPolicyTestFunc func(
	t *testing.T,
	policyBranch string,
	returnJSON bool,
	policyTreeDepth uint,
	sizeLimit uint,
) ([]byte, error)

type mockPolicyClient struct {
	t            *testing.T
	loadPolicy   loadPolicyTestFunc
	dryRunPolicy dryRunPolicyTestFunc
	fetchPolicy  fetchPolicyTestFunc
}

func (m mockPolicyClient) LoadPolicy(
	mode conjurapi.PolicyMode,
	policyBranch string,
	policySrc io.Reader,
) (*conjurapi.PolicyResponse, error) {
	return m.loadPolicy(m.t, mode, policyBranch, policySrc)
}

func (m mockPolicyClient) DryRunPolicy(
	mode conjurapi.PolicyMode,
	policyBranch string,
	policySrc io.Reader,
) (*conjurapi.DryRunPolicyResponse, error) {
	return m.dryRunPolicy(m.t, mode, policyBranch, policySrc)
}

func (m mockPolicyClient) FetchPolicy(
	policyBranch string,
	returnJSON bool,
	policyTreeDepth uint,
	sizeLimit uint,
) ([]byte, error) {
	return m.fetchPolicy(m.t, policyBranch, returnJSON, policyTreeDepth, sizeLimit)
}

type policyCmdTestCase struct {
	name string
	args []string // $TMPDIR or $TMPFILE in any of the args is substituted
	// for the temporary directory/file created for each test
	beforeTest         func(t *testing.T, pathToTmpfile string)
	loadPolicy         loadPolicyTestFunc
	dryRunPolicy       dryRunPolicyTestFunc
	fetchPolicy        fetchPolicyTestFunc
	promptResponses    []promptResponse
	clientFactoryError error
	assert             func(t *testing.T, stdout string, stderr string, err error, pathToTmpDir string)
}

var policyCmdTestCases = []policyCmdTestCase{
	{
		name: "policy command help",
		args: []string{"policy", "--help"},
		assert: func(t *testing.T, stdout, stderr string, err error, pathToTmpDir string) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "fetch subcommand help",
		args: []string{"policy", "fetch", "--help"},
		assert: func(t *testing.T, stdout, stderr string, err error, pathToTmpDir string) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name:               "fetch subcommand client factory error",
		args:               []string{"policy", "fetch", "-b", "meow"},
		clientFactoryError: fmt.Errorf("client factory error"),
		assert: func(t *testing.T, stdout, stderr string, err error, pathToTmpDir string) {
			assert.Contains(t, stderr, "Error: client factory error")
		},
	},
	{
		name: "fetch subcommand missing branch",
		args: []string{"policy", "fetch"},
		assert: func(t *testing.T, stdout, stderr string, err error, pathToTmpDir string) {
			assert.Contains(t, stderr, "Error: required flag(s) \"branch\" not set\n")
		},
	},
	{
		name: "fetch subcommand invalid output type",
		args: []string{"policy", "fetch", "-b", "meow", "-o", "xml"},
		assert: func(t *testing.T, stdout, stderr string, err error, pathToTmpDir string) {
			assert.Contains(t, stderr, "Error: output format must be 'yaml' or 'json'\n")
		},
	},
	{
		name: "fetch subcommand invalid file path",
		args: []string{"policy", "fetch", "-b", "meow", "-f", "/non/existent/path/somefile.yaml"},
		fetchPolicy: func(
			t *testing.T,
			policyBranch string,
			returnJSON bool,
			policyTreeDepth uint,
			sizeLimit uint,
		) ([]byte, error) {
			return []byte("---\npolicy\n  id: root\n  body: []"), nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error, pathToTmpDir string) {
			assert.Contains(t, stderr, "Error: directory /non/existent/path does not exist\n")
		},
	},
	{
		name: "fetch subcommand with good YAML response",
		args: []string{"policy", "fetch", "-b", "meow"},
		fetchPolicy: func(
			t *testing.T,
			policyBranch string,
			returnJSON bool,
			policyTreeDepth uint,
			sizeLimit uint,
		) ([]byte, error) {
			return []byte("---\npolicy\n  id: root\n  body: []"), nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error, pathToTmpDir string) {
			assert.Equal(t, "---\npolicy\n  id: root\n  body: []\n", stdout)
		},
	},
	{
		name: "fetch subcommand with good JSON response",
		args: []string{"policy", "fetch", "-b", "meow", "-o", "json"},
		fetchPolicy: func(
			t *testing.T,
			policyBranch string,
			returnJSON bool,
			policyTreeDepth uint,
			sizeLimit uint,
		) ([]byte, error) {
			return []byte("[{\"policy\":{\"id\":\"root\",\"body\":[]}}]"), nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error, pathToTmpDir string) {
			json, _ := utils.PrettyPrintJSON([]byte("[{\"policy\":{\"id\":\"root\",\"body\":[]}}]\n"))
			assert.Equal(t, string(json), stdout)
		},
	},
	{
		name: "fetch subcommand with good YAML response saved to a file",
		args: []string{"policy", "fetch", "-b", "meow", "-f", "$TMPDIR/somefile.yaml"},
		fetchPolicy: func(
			t *testing.T,
			policyBranch string,
			returnJSON bool,
			policyTreeDepth uint,
			sizeLimit uint,
		) ([]byte, error) {
			return []byte("---\npolicy\n  id: root\n  body: []"), nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error, pathToTmpDir string) {
			assert.Empty(t, stdout)
			contents, _ := os.ReadFile(pathToTmpDir + "/somefile.yaml")
			assert.Equal(t, "---\npolicy\n  id: root\n  body: []", string(contents))
		},
	},
	{
		name: "fetch subcommand response error",
		args: []string{"policy", "fetch", "-b", "meow"},
		fetchPolicy: func(
			t *testing.T,
			policyBranch string,
			returnJSON bool,
			policyTreeDepth uint,
			sizeLimit uint,
		) ([]byte, error) {
			return nil, fmt.Errorf("%s", "some error")
		},
		assert: func(t *testing.T, stdout, stderr string, err error, pathToTmpDir string) {
			assert.Contains(t, stderr, "Error: some error")
		},
	},
}

func sharedLoadPolicyCmdTestCases(
	subcommand string,
	expectedMode conjurapi.PolicyMode,
) []policyCmdTestCase {
	return []policyCmdTestCase{
		{
			name: fmt.Sprintf("%s subcommand help", subcommand),
			args: []string{"policy", subcommand, "--help"},
			assert: func(t *testing.T, stdout, stderr string, err error, pathToTmpDir string) {
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
			assert: func(t *testing.T, stdout, stderr string, err error, pathToTmpDir string) {
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
			assert: func(t *testing.T, stdout, stderr string, err error, pathToTmpDir string) {
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
			assert: func(t *testing.T, stdout, stderr string, err error, pathToTmpDir string) {
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
			assert: func(t *testing.T, stdout, stderr string, err error, pathToTmpDir string) {
				assert.Contains(t, stderr, "Error: some error")
			},
		},
		{
			name:               fmt.Sprintf("%s subcommand client factory error", subcommand),
			args:               []string{"policy", subcommand, "-b", "meow", "-f", "-"},
			clientFactoryError: fmt.Errorf("client factory error"),
			assert: func(t *testing.T, stdout, stderr string, err error, pathToTmpDir string) {
				assert.Contains(t, stderr, "Error: client factory error")
			},
		},
		{
			name: fmt.Sprintf("%s subcommand missing file", subcommand),
			args: []string{"policy", subcommand, "-b", "meow"},
			assert: func(t *testing.T, stdout, stderr string, err error, pathToTmpDir string) {
				assert.Contains(t, stderr, "Error: required flag(s) \"file\" not set\n")
			},
		},
		{
			name: fmt.Sprintf("%s subcommand missing branch", subcommand),
			args: []string{"policy", subcommand, "-f", "-"},
			assert: func(t *testing.T, stdout, stderr string, err error, pathToTmpDir string) {
				assert.Contains(t, stderr, "Error: required flag(s) \"branch\" not set\n")
			},
		},
		{
			name: fmt.Sprintf("%s subcommand with good response (dryrun)", subcommand),
			args: []string{"policy", subcommand, "-b", "meow", "--dry-run", "-f", "-"},
			dryRunPolicy: func(
				t *testing.T,
				mode conjurapi.PolicyMode,
				policyBranch string,
				policySrc io.Reader,
			) (*conjurapi.DryRunPolicyResponse, error) {
				return &conjurapi.DryRunPolicyResponse{
					Status: "Valid YAML",
				}, nil
			},
			assert: func(t *testing.T, stdout, stderr string, err error, pathToTmpDir string) {
				assert.Contains(t, stdout, "Valid YAML")
				assert.Contains(t, stderr, "Dry run policy 'meow'")
			},
		},
		{
			name: fmt.Sprintf("%s subcommand with bad response (dryrun)", subcommand),
			args: []string{"policy", subcommand, "-b", "meow", "--dry-run", "-f", "-"},
			dryRunPolicy: func(
				t *testing.T,
				mode conjurapi.PolicyMode,
				policyBranch string,
				policySrc io.Reader,
			) (*conjurapi.DryRunPolicyResponse, error) {
				return &conjurapi.DryRunPolicyResponse{
					Status: "Invalid YAML",
					Errors: []conjurapi.DryRunErrors{
						{
							Line:    0,
							Column:  0,
							Message: "undefined method `referenced_records' for \"user alice\":String\n",
						},
					},
				}, nil
			},
			assert: func(t *testing.T, stdout, stderr string, err error, pathToTmpDir string) {
				assert.Contains(t, stdout, "status")
				assert.Contains(t, stdout, "errors")
				assert.Contains(t, stdout, "line")
				assert.Contains(t, stdout, "column")
				assert.Contains(t, stdout, "message")
				assert.Contains(t, stdout, "Invalid YAML")
				assert.Contains(t, stderr, "Dry run policy 'meow'")
			},
		},
	}
}

func TestPolicyCmd(t *testing.T) {
	t.Parallel()

	var allTests []policyCmdTestCase
	for _, cases := range [][]policyCmdTestCase{
		policyCmdTestCases,
		sharedLoadPolicyCmdTestCases(
			"load",
			conjurapi.PolicyModePost,
		),
		sharedLoadPolicyCmdTestCases(
			"update",
			conjurapi.PolicyModePatch,
		),
		sharedLoadPolicyCmdTestCases(
			"replace",
			conjurapi.PolicyModePut,
		),
	} {
		allTests = append(allTests, cases...)
	}

	for _, tc := range allTests {
		t.Run(tc.name, func(t *testing.T) {
			pathToTmpDir := t.TempDir()
			pathToTmpfile := pathToTmpDir + "/file"
			os.Remove(pathToTmpfile)
			defer os.Remove(pathToTmpfile)

			if tc.beforeTest != nil {
				tc.beforeTest(t, pathToTmpfile)
			}

			mockClient := mockPolicyClient{t: t, fetchPolicy: tc.fetchPolicy, loadPolicy: tc.loadPolicy, dryRunPolicy: tc.dryRunPolicy}

			cmd := newPolicyCommand(
				func(cmd *cobra.Command) (policyClient, error) {
					return mockClient, tc.clientFactoryError
				},
			)

			// $TMPFILE points to tempfile created for each test run
			for i, v := range tc.args {
				args := strings.Replace(v, "$TMPDIR", pathToTmpDir, 1)
				tc.args[i] = strings.Replace(args, "$TMPFILE", pathToTmpfile, 1)
			}

			if tc.promptResponses != nil {
				// Use the prompt responses helper to simulate user input
				stdout, err := executeCommandForTestWithPromptResponses(
					t, cmd, tc.promptResponses,
				)
				if tc.assert != nil {
					tc.assert(t, stdout, "", err, pathToTmpDir)
				}
			} else {
				stdout, stderr, err := executeCommandForTest(
					t, cmd, tc.args...,
				)
				if tc.assert != nil {
					tc.assert(t, stdout, stderr, err, pathToTmpDir)
				}
			}
		})
	}
}
