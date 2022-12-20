package cmd

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type mockWhoamiClient struct {
	whoami func() ([]byte, error)
}

func (m mockWhoamiClient) WhoAmI() ([]byte, error) {
	return m.whoami()
}

var whoamiCmdTestCases = []struct {
	name               string
	args               []string
	whoami             func() ([]byte, error)
	clientFactoryError error
	assert             func(t *testing.T, stdout, stderr string, err error)
}{
	{
		name: "display help",
		args: []string{"whoami", "--help"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "json response",
		args: []string{"whoami"},
		whoami: func() ([]byte, error) {
			return []byte(`{"user":"test"}`), nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			expectedOut := `{
  "user": "test"
}
`
			assert.Contains(t, stdout, expectedOut)
		},
	},
	{
		name: "non-json response",
		args: []string{"whoami"},
		whoami: func() ([]byte, error) {
			return []byte(`use is test`), nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: invalid")
		},
	},
	{
		name: "client error",
		args: []string{"whoami"},
		whoami: func() ([]byte, error) {
			return nil, fmt.Errorf("%s", "an error")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: an error\n")
		},
	},
	{
		name:               "client factory error",
		args:               []string{"whoami"},
		clientFactoryError: fmt.Errorf("%s", "client factory error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: client factory error\n")
		},
	},
}

func TestWhoamiCmd(t *testing.T) {
	t.Parallel()

	for _, tc := range whoamiCmdTestCases {
		t.Run(tc.name, func(t *testing.T) {
			testWhoamiClientFactory := func(cmd *cobra.Command) (whoamiClient, error) {
				return mockWhoamiClient{whoami: tc.whoami}, tc.clientFactoryError
			}
			cmd := newWhoamiCommand(testWhoamiClientFactory)
			stdout, stderr, err := executeCommandForTest(t, cmd, tc.args...)
			tc.assert(t, stdout, stderr, err)
		})
	}
}
