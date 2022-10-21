package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
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

var policyCmdTestCases = []struct {
	name               string
	args               []string
	beforeTest         func(t *testing.T, pathToTmpfile string)
	loadPolicy         loadPolicyTestFunc
	promptResponses    []promptResponse
	clientFactoryError error
	assert             func(t *testing.T, stdout string, stderr string, err error)
}{
	{
		name: "policy command help",
		args: []string{"policy", "--help"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "Use the policy command to manage Conjur policies")
		},
	},
	{
		name: "load subcommand help",
		args: []string{"policy", "load", "--help"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "Load a policy and create resources")
		},
	},
	{
		name: "replace subcommand help",
		args: []string{"policy", "replace", "--help"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "Fully replace an existing policy")
		},
	},
	{
		name: "append subcommand help",
		args: []string{"policy", "append", "--help"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "Update existing resources in the policy or create new resources")
		},
	},
	{
		name: "load subcommand",
		args: []string{"policy", "load", "-b", "meow", "-f", "-"},
		loadPolicy: func(
			t *testing.T,
			mode conjurapi.PolicyMode,
			policyBranch string,
			policySrc io.Reader,
		) (*conjurapi.PolicyResponse, error) {
			// Assert on mode
			assert.Equal(t, conjurapi.PolicyModePost, mode)

			return nil, nil
		},
	},
	{
		name: "load subcommand with good response",
		args: []string{"policy", "load", "-b", "meow", "-f", "-"},
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
		name: "load subcommand from stdin",
		promptResponses: []promptResponse{
			{
				response: "policy file content\n",
			},
		},
		args: []string{"policy", "load", "-b", "meow", "-f", "-"},
		loadPolicy: func(
			t *testing.T,
			mode conjurapi.PolicyMode,
			policyBranch string,
			policySrc io.Reader,
		) (*conjurapi.PolicyResponse, error) {
			policyContents, err := ioutil.ReadAll(policySrc)
			assert.NoError(t, err)
			assert.Equal(t, "policy file content\n", string(policyContents))

			return nil, nil
		},
	},
	{
		name: "load subcommand from file",
		args: []string{"policy", "load", "-b", "meow", "-f", "TMPFILE"},
		beforeTest: func(t *testing.T, pathToTmpfile string) {
			err := ioutil.WriteFile(pathToTmpfile, []byte("policy file content"), 0644)
			assert.NoError(t, err)
		},
		loadPolicy: func(
			t *testing.T,
			mode conjurapi.PolicyMode,
			policyBranch string,
			policySrc io.Reader,
		) (*conjurapi.PolicyResponse, error) {
			policyContents, err := ioutil.ReadAll(policySrc)
			assert.NoError(t, err)
			assert.Equal(t, "policy file content", string(policyContents))

			return nil, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.NoError(t, err)
		},
	},
	{
		name: "load subcommand response error",
		args: []string{"policy", "load", "-b", "meow", "-f", "-"},
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
}

func TestPolicyCmd(t *testing.T) {
	t.Parallel()

	for _, tc := range policyCmdTestCases {
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

			for i, v := range tc.args {
				tc.args[i] = strings.Replace(v, "TMPFILE", pathToTmpfile, 1)
			}

			stdout, stderr, err := executeCommandForTestWithPromptResponses(
				t, cmd, tc.promptResponses, tc.args...,
			)
			if tc.assert != nil {
				tc.assert(t, stdout, stderr, err)
			}
		})
	}
}
