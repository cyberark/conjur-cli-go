package cmd

import (
	"fmt"
	"testing"

	"github.com/cyberark/conjur-cli-go/pkg/test"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type mockWhoamiClient struct {
	whoami func() ([]byte, error)
}

func (m mockWhoamiClient) WhoAmI() ([]byte, error) {
	return m.whoami()
}

func TestWhoamiCmd(t *testing.T) {
	tt := []struct {
		name               string
		args               []string
		whoami             func() ([]byte, error)
		clientFactoryError error
		out                string
	}{
		{
			name: "display help",
			args: []string{"--help"},
			out:  "Displays info about the logged in user",
		},
		{
			name: "json response",
			args: []string{},
			whoami: func() ([]byte, error) {
				return []byte(`{"user":"test"}`), nil
			},
			out: `{
  "user": "test"
}
`,
		},
		{
			name: "non-json response",
			args: []string{},
			whoami: func() ([]byte, error) {
				return []byte(`use is test`), nil
			},
			out: `Error: invalid`,
		},
		{
			name: "error",
			args: []string{},
			whoami: func() ([]byte, error) {
				return nil, fmt.Errorf("%s", "an error")
			},
			out: "Error: an error\n",
		},
		{
			name:               "client factory error",
			args:               []string{},
			clientFactoryError: fmt.Errorf("%s", "client factory error"),
			out:                "Error: client factory error\n",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			testWhoamiClientFactory := func(cmd *cobra.Command) (whoamiClient, error) {
				return mockWhoamiClient{whoami: tc.whoami}, tc.clientFactoryError
			}
			cmd := NewWhoamiCommand(testWhoamiClientFactory)

			out, _ := test.Execute(t, cmd, tc.args...)
			assert.Contains(t, out, tc.out)
		})
	}
}
