package cmd

import (
	"fmt"
	"testing"

	"github.com/cyberark/conjur-cli-go/pkg/testutils"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type mockVariableClient struct {
	t   *testing.T
	get func(*testing.T, string) ([]byte, error)
	set func(*testing.T, string, string) error
}

func (m mockVariableClient) RetrieveSecret(path string) ([]byte, error) {
	return m.get(m.t, path)
}
func (m mockVariableClient) AddSecret(path string, value string) error {
	return m.set(m.t, path, value)
}

var variableCmdTestCases = []struct {
	name               string
	args               []string
	get                func(t *testing.T, path string) ([]byte, error)
	set                func(t *testing.T, path string, value string) error
	clientFactoryError error
	assert             func(t *testing.T, stdout string, stderr string, err error)
}{
	{
		name: "variable command help",
		args: []string{"--help"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "Use the variable command to manage Conjur variables")
		},
	},
	{
		name: "get subcommand help",
		args: []string{"get", "--help"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "Use the get subcommand to get the value of one or more Conjur variables")
		},
	},
	{
		name: "set subcommand help",
		args: []string{"set", "--help"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "Use the set subcommand to set the value of a Conjur variable")
		},
	},
	{
		name: "get subcommand",
		args: []string{"get", "-i", "meow"},
		get: func(t *testing.T, path string) ([]byte, error) {
			// Assert on arguments
			assert.Equal(t, "meow", path)

			return []byte("moo"), nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "moo")
		},
	},
	{
		name: "get subcommand error",
		args: []string{"get", "-i", "meow"},
		get: func(t *testing.T, path string) ([]byte, error) {
			return nil, fmt.Errorf("%s", "get error")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: get error")
		},
	},
	{
		name: "get subcommand missing required flags",
		args: []string{"get"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: required flag(s) \"id\" not set\n")
		},
	},
	{
		name:               "get client factory error",
		args:               []string{"get", "-i", "moo"},
		clientFactoryError: fmt.Errorf("%s", "client factory error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: client factory error\n")
		},
	},
	{
		name: "set subcommand",
		args: []string{"set", "-i", "meow", "-v", "moo"},
		set: func(t *testing.T, path, value string) error {
			// Assert on arguments
			assert.Equal(t, "meow", path)
			assert.Equal(t, "moo", value)

			return nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "Value added")
		},
	},
	{
		name: "set subcommand error",
		args: []string{"set", "-i", "meow", "-v", "moo"},
		set: func(t *testing.T, path, value string) error {
			return fmt.Errorf("%s", "set error")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: set error")
		},
	},
	{
		name: "set subcommand missing required flags",
		args: []string{"set"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: required flag(s) \"id\", \"value\" not set\n")
		},
	},
	{
		name:               "set client factory error",
		args:               []string{"set", "-i", "meow", "-v", "moo"},
		clientFactoryError: fmt.Errorf("%s", "client factory error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: client factory error\n")
		},
	},
}

func TestVariableCmd(t *testing.T) {
	for _, tc := range variableCmdTestCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := mockVariableClient{t: t, set: tc.set, get: tc.get}

			cmd := newVariableCmd(
				func(cmd *cobra.Command) (variableGetClient, error) {
					return mockClient, tc.clientFactoryError
				},
				func(cmd *cobra.Command) (variableSetClient, error) {
					return mockClient, tc.clientFactoryError
				},
			)

			stdout, stderr, err := testutils.Execute(t, cmd, tc.args...)
			tc.assert(t, stdout, stderr, err)
		})
	}
}
