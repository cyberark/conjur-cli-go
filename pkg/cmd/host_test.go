package cmd

import (
	"fmt"
	"testing"

	"github.com/cyberark/conjur-cli-go/pkg/testutils"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type mockHostRotateAPIKeyClient struct {
	hostRotateAPIKey func(roleID string) ([]byte, error)
}

func (m mockHostRotateAPIKeyClient) RotateAPIKey(roleID string) ([]byte, error) {
	return m.hostRotateAPIKey(roleID)
}

var hostRotateAPIKeyCmdTestCases = []struct {
	name               string
	args               []string
	hostRotateAPIKey   func(roleID string) ([]byte, error)
	clientFactoryError error
	assert             func(t *testing.T, stdout, stderr string, err error)
}{
	{
		name: "display help",
		args: []string{"--help"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "Rotate a host's API key")
		},
	},
	{
		name: "successful rotation",
		args: []string{"--host=dev-host"},
		hostRotateAPIKey: func(roleID string) ([]byte, error) {
			return []byte("test-api-key"), nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "test-api-key")
		},
	},
	{
		name: "client error",
		args: []string{"--host=dev-host"},
		hostRotateAPIKey: func(roleID string) ([]byte, error) {
			return nil, fmt.Errorf("%s", "an error")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: an error\n")
		},
	},
	{
		name:               "client factory error",
		args:               []string{"--host=dev-host"},
		clientFactoryError: fmt.Errorf("%s", "client factory error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: client factory error\n")
		},
	},
}

func TestHostRotateAPIKeyCmd(t *testing.T) {
	for _, tc := range hostRotateAPIKeyCmdTestCases {
		t.Run(tc.name, func(t *testing.T) {
			testHostRotateAPIKeyClientFactory := func(cmd *cobra.Command) (hostRotateAPIKeyClient, error) {
				return mockHostRotateAPIKeyClient{hostRotateAPIKey: tc.hostRotateAPIKey}, tc.clientFactoryError
			}
			cmd := NewHostRotateAPIKeyCommand(testHostRotateAPIKeyClientFactory)

			stdout, stderr, err := testutils.Execute(t, cmd, tc.args...)
			tc.assert(t, stdout, stderr, err)
		})
	}
}
