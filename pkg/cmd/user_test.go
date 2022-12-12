package cmd

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type mockUserClient struct {
	t                *testing.T
	userRotateAPIKey func(*testing.T, string) ([]byte, error)
	whoAmI           func() ([]byte, error)
}

func (m mockUserClient) RotateUserAPIKey(userID string) ([]byte, error) {
	return m.userRotateAPIKey(m.t, userID)
}

func (m mockUserClient) WhoAmI() ([]byte, error) {
	return m.whoAmI()
}

var userRotateAPIKeyCmdTestCases = []struct {
	name               string
	args               []string
	userRotateAPIKey   func(t *testing.T, userID string) ([]byte, error)
	whoAmI             func() ([]byte, error)
	clientFactoryError error
	assert             func(t *testing.T, stdout, stderr string, err error)
}{
	{
		name: "display help",
		args: []string{"user", "rotate-api-key", "--help"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "rotate API key for logged in user",
		args: []string{"user", "rotate-api-key"},
		userRotateAPIKey: func(t *testing.T, userID string) ([]byte, error) {
			// Assert on arguments
			assert.Equal(t, "logged-in-user", userID)

			return []byte("test-api-key"), nil
		},
		whoAmI: func() ([]byte, error) {
			responseStr := "{\"client_ip\":\"12.16.23.10\",\"user_agent\":\"curl/7.64.1\",\"account\":\"demo\",\"username\":\"logged-in-user\",\"token_issued_at\":\"2020-09-14T19:50:42.000+00:00\"}"
			responseBytes := []byte(responseStr)
			return responseBytes, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "test-api-key")
		},
	},
	{
		name: "rotate API key for specified user",
		args: []string{"user", "rotate-api-key", "--user-id=dev-user"},
		userRotateAPIKey: func(t *testing.T, userID string) ([]byte, error) {
			// Assert on arguments
			assert.Equal(t, "dev-user", userID)

			return []byte("test-api-key"), nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "test-api-key")
		},
	},
	{
		name: "client error when rotating API key",
		args: []string{"user", "rotate-api-key", "--user-id=dev-user"},
		userRotateAPIKey: func(t *testing.T, userID string) ([]byte, error) {
			return nil, fmt.Errorf("%s", "an error")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: an error\n")
		},
	},
	{
		name: "client error when retrieving username of logged in user",
		args: []string{"user", "rotate-api-key"},
		whoAmI: func() ([]byte, error) {
			return nil, fmt.Errorf("%s", "an error")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: an error\n")
		},
	},
	{
		name:               "client factory error",
		args:               []string{"user", "rotate-api-key", "--user-id=dev-user"},
		clientFactoryError: fmt.Errorf("%s", "client factory error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: client factory error\n")
		},
	},
}

func TestUserRotateAPIKeyCmd(t *testing.T) {
	for _, tc := range userRotateAPIKeyCmdTestCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := mockUserClient{
				t:                t,
				userRotateAPIKey: tc.userRotateAPIKey,
				whoAmI:           tc.whoAmI,
			}

			cmd := newUserCmd(
				func(cmd *cobra.Command) (userClient, error) {
					return mockClient, tc.clientFactoryError
				},
			)

			stdout, stderr, err := executeCommandForTest(t, cmd, tc.args...)

			tc.assert(t, stdout, stderr, err)
		})
	}
}
