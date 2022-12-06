package cmd

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

const _hostRotateAPIKeyHelpStr = "Rotate the API key of the host specified by the [host-id] parameter."

type mockHostClient struct {
	t                *testing.T
	hostRotateAPIKey func(*testing.T, string) ([]byte, error)
}

func (m mockHostClient) RotateHostAPIKey(hostID string) ([]byte, error) {
	return m.hostRotateAPIKey(m.t, hostID)
}

var hostRotateAPIKeyCmdTestCases = []struct {
	name               string
	args               []string
	hostRotateAPIKey   func(t *testing.T, hostID string) ([]byte, error)
	clientFactoryError error
	assert             func(t *testing.T, stdout, stderr string, err error)
}{
	{
		name: "display help",
		args: []string{"host", "rotate-api-key", "--help"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, _hostRotateAPIKeyHelpStr)
		},
	},
	{
		name: "successful rotation",
		args: []string{"host", "rotate-api-key", "--host-id=dev-host"},
		hostRotateAPIKey: func(t *testing.T, hostID string) ([]byte, error) {
			// Assert on arguments
			assert.Equal(t, "dev-host", hostID)

			return []byte("test-api-key"), nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "test-api-key")
		},
	},
	{
		name: "client error",
		args: []string{"host", "rotate-api-key", "--host-id=dev-host"},
		hostRotateAPIKey: func(t *testing.T, hostID string) ([]byte, error) {
			return nil, fmt.Errorf("%s", "an error")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: an error\n")
		},
	},
	{
		name:               "client factory error",
		args:               []string{"host", "rotate-api-key", "--host-id=dev-host"},
		clientFactoryError: fmt.Errorf("%s", "client factory error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: client factory error\n")
		},
	},
}

func TestHostRotateAPIKeyCmd(t *testing.T) {
	for _, tc := range hostRotateAPIKeyCmdTestCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := mockHostClient{t: t, hostRotateAPIKey: tc.hostRotateAPIKey}

			cmd := newHostCmd(
				func(cmd *cobra.Command) (hostClient, error) {
					return mockClient, tc.clientFactoryError
				},
			)

			stdout, stderr, err := executeCommandForTest(t, cmd, tc.args...)
			tc.assert(t, stdout, stderr, err)
		})
	}
}
