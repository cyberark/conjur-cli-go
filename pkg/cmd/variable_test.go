package cmd

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type mockVariableClient struct {
	t              *testing.T
	get            func(*testing.T, string) ([]byte, error)
	getWithVersion func(*testing.T, string, int) ([]byte, error)
	getBatch       func(*testing.T, []string) (map[string][]byte, error)
	set            func(*testing.T, string, string) error
}

func (m mockVariableClient) RetrieveSecret(path string) ([]byte, error) {
	return m.get(m.t, path)
}
func (m mockVariableClient) RetrieveSecretWithVersion(path string, version int) ([]byte, error) {
	return m.getWithVersion(m.t, path, version)
}
func (m mockVariableClient) RetrieveBatchSecretsSafe(paths []string) (map[string][]byte, error) {
	return m.getBatch(m.t, paths)
}
func (m mockVariableClient) AddSecret(path string, value string) error {
	return m.set(m.t, path, value)
}

var variableCmdTestCases = []struct {
	name               string
	args               []string
	get                func(t *testing.T, path string) ([]byte, error)
	getWithVersion     func(t *testing.T, path string, version int) ([]byte, error)
	getBatch           func(t *testing.T, paths []string) (map[string][]byte, error)
	set                func(t *testing.T, path string, value string) error
	clientFactoryError error
	assert             func(t *testing.T, stdout string, stderr string, err error)
}{
	{
		name: "variable command help",
		args: []string{"variable", "--help"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "get subcommand help",
		args: []string{"variable", "get", "--help"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "set subcommand help",
		args: []string{"variable", "set", "--help"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "get subcommand",
		args: []string{"variable", "get", "-i", "meow"},
		getBatch: func(t *testing.T, path []string) (map[string][]byte, error) {
			// Assert on arguments
			assert.Equal(t, ("meow"), path[0])

			return map[string][]byte{"meow": []byte("moo")}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "moo")
		},
	},
	{
		name: "get subcommand error",
		args: []string{"variable", "get", "-i", "meow"},
		getBatch: func(t *testing.T, path []string) (map[string][]byte, error) {
			return nil, fmt.Errorf("%s", "get error")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: get error")
		},
	},
	{
		name: "get subcommand missing required flags",
		args: []string{"variable", "get"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: required flag(s) \"id\" not set\n")
		},
	},
	{
		name: "get with version",
		args: []string{"variable", "get", "-i", "moo", "-v", "1"},
		getWithVersion: func(t *testing.T, path string, version int) ([]byte, error) {
			// Assert on arguments
			assert.Equal(t, 1, version)

			return []byte("moo"), nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "moo")
		},
	},
	{
		name:               "get client factory error",
		args:               []string{"variable", "get", "-i", "moo"},
		clientFactoryError: fmt.Errorf("%s", "client factory error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: client factory error\n")
		},
	},
	{
		name: "get two variables",
		args: []string{"variable", "get", "-i", "meow,woof"},
		getBatch: func(t *testing.T, path []string) (map[string][]byte, error) {
			// Assert on arguments
			assert.Equal(t, "meow", path[0])
			assert.Equal(t, "woof", path[1])
			return map[string][]byte{"meow": []byte("moo"), "woof": []byte("quack")}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "moo")
			assert.Contains(t, stdout, "quack")
		},
	},
	{
		name: "get two variables with version",
		args: []string{"variable", "get", "-i", "meow,woof", "-v", "1"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "version can not be used with multiple variables")
		},
	},
	{
		name: "set subcommand",
		args: []string{"variable", "set", "-i", "meow", "-v", "moo"},
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
		args: []string{"variable", "set", "-i", "meow", "-v", "moo"},
		set: func(t *testing.T, path, value string) error {
			return fmt.Errorf("%s", "set error")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: set error")
		},
	},
	{
		name: "set subcommand missing required flags",
		args: []string{"variable", "set"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: required flag(s) \"id\", \"value\" not set\n")
		},
	},
	{
		name:               "set client factory error",
		args:               []string{"variable", "set", "-i", "meow", "-v", "moo"},
		clientFactoryError: fmt.Errorf("%s", "client factory error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: client factory error\n")
		},
	},
}

func TestVariableCmd(t *testing.T) {
	t.Parallel()

	for _, tc := range variableCmdTestCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := mockVariableClient{
				t: t, set: tc.set, get: tc.get, getWithVersion: tc.getWithVersion, getBatch: tc.getBatch,
			}

			cmd := newVariableCmd(
				func(cmd *cobra.Command) (variableGetClient, error) {
					return mockClient, tc.clientFactoryError
				},
				func(cmd *cobra.Command) (variableSetClient, error) {
					return mockClient, tc.clientFactoryError
				},
			)

			stdout, stderr, err := executeCommandForTest(t, cmd, tc.args...)
			tc.assert(t, stdout, stderr, err)
		})
	}
}
