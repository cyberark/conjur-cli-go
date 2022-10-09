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

func TestVariableCmd(t *testing.T) {
	tt := []struct {
		name               string
		args               []string
		get                func(t *testing.T, path string) ([]byte, error)
		set                func(t *testing.T, path string, value string) error
		clientFactoryError error
		out                string
	}{
		{
			name: "variable command help",
			args: []string{"--help"},
			out:  "Use the variable command to manage Conjur variables",
		},
		{
			name: "get subcommand help",
			args: []string{"get", "--help"},
			out:  "Use the get subcommand to get the value of one or more Conjur variables",
		},
		{
			name: "set subcommand help",
			args: []string{"set", "--help"},
			out:  "Use the set subcommand to set the value of a Conjur variable",
		},
		{
			name: "get subcommand",
			args: []string{"get", "-i", "meow"},
			get: func(t *testing.T, path string) ([]byte, error) {
				// Assert on arguments
				assert.Equal(t, "meow", path)

				return []byte("moo"), nil
			},
			out: "moo",
		},
		{
			name: "get subcommand error",
			args: []string{"get", "-i", "meow"},
			get: func(t *testing.T, path string) ([]byte, error) {
				return nil, fmt.Errorf("%s", "get error")
			},
			out: "Error: get error",
		},
		{
			name: "get subcommand missing required flags",
			args: []string{"get"},
			out:  "Error: required flag(s) \"id\" not set\n",
		},
		{
			name:               "get client factory error",
			args:               []string{"get", "-i", "moo"},
			clientFactoryError: fmt.Errorf("%s", "client factory error"),
			out:                "Error: client factory error\n",
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
			out: "Value added",
		},
		{
			name: "set subcommand error",
			args: []string{"set", "-i", "meow", "-v", "moo"},
			set: func(t *testing.T, path, value string) error {
				return fmt.Errorf("%s", "set error")
			},
			out: "Error: set error",
		},
		{
			name: "set subcommand missing required flags",
			args: []string{"set"},
			out:  "Error: required flag(s) \"id\", \"value\" not set\n",
		},
		{
			name:               "set client factory error",
			args:               []string{"set", "-i", "meow", "-v", "moo"},
			clientFactoryError: fmt.Errorf("%s", "client factory error"),
			out:                "Error: client factory error\n",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := mockVariableClient{t: t, set: tc.set, get: tc.get}

			cmd := NewVariableCmd(
				func(cmd *cobra.Command) (variableGetClient, error) {
					return mockClient, tc.clientFactoryError
				},
				func(cmd *cobra.Command) (variableSetClient, error) {
					return mockClient, tc.clientFactoryError
				},
			)

			out, _ := testutils.Execute(t, cmd, tc.args...)
			assert.Contains(t, out, tc.out)
		})
	}
}
