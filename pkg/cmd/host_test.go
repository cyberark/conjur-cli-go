package cmd

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type mockHostClient struct {
	t                       *testing.T
	rotateCurrentUserAPIKey func() ([]byte, error)
	rotateHostAPIKey        func(*testing.T, string) ([]byte, error)
}

func (m mockHostClient) RotateCurrentUserAPIKey() ([]byte, error) {
	return m.rotateCurrentUserAPIKey()
}

func (m mockHostClient) RotateHostAPIKey(hostID string) ([]byte, error) {
	return m.rotateHostAPIKey(m.t, hostID)
}

var rotateHostAPIKeyCmdTestCases = []struct {
	name                    string
	args                    []string
	rotateCurrentUserAPIKey func() ([]byte, error)
	rotateHostAPIKey        func(t *testing.T, hostID string) ([]byte, error)
	clientFactoryError      error
	assert                  func(t *testing.T, stdout, stderr string, err error)
}{
	{
		name: "without subcommand",
		args: []string{"host"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "display help",
		args: []string{"host", "rotate-api-key", "--help"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "rotate API key for logged in host",
		args: []string{"host", "rotate-api-key"},
		rotateCurrentUserAPIKey: func() ([]byte, error) {
			return []byte("test-api-key"), nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "test-api-key")
		},
	},
	{
		name: "rotate API key for specified host",
		args: []string{"host", "rotate-api-key", "--id=dev-host"},
		rotateHostAPIKey: func(t *testing.T, hostID string) ([]byte, error) {
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
		args: []string{"host", "rotate-api-key", "--id=dev-host"},
		rotateHostAPIKey: func(t *testing.T, hostID string) ([]byte, error) {
			return nil, fmt.Errorf("%s", "an error")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: an error\n")
		},
	},
	{
		name:               "client factory error",
		args:               []string{"host", "rotate-api-key", "--id=dev-host"},
		clientFactoryError: fmt.Errorf("%s", "client factory error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: client factory error\n")
		},
	},
}

func TestRotateHostAPIKeyCmd(t *testing.T) {
	for _, tc := range rotateHostAPIKeyCmdTestCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := mockHostClient{t: t, rotateHostAPIKey: tc.rotateHostAPIKey, rotateCurrentUserAPIKey: tc.rotateCurrentUserAPIKey}

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
