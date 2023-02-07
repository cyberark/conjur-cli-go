package cmd

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type mockUserClient struct {
	t                         *testing.T
	whoAmI                    func() ([]byte, error)
	rotateUserAPIKey          func(*testing.T, string) ([]byte, error)
	changeCurrentUserPassword func(*testing.T, string) ([]byte, error)
}

func (m mockUserClient) WhoAmI() ([]byte, error) {
	return m.whoAmI()
}

func (m mockUserClient) RotateUserAPIKey(userID string) ([]byte, error) {
	return m.rotateUserAPIKey(m.t, userID)
}

func (m mockUserClient) ChangeCurrentUserPassword(newPassword string) ([]byte, error) {
	return m.changeCurrentUserPassword(m.t, newPassword)
}

var userRotateAPIKeyCmdTestCases = []struct {
	name               string
	args               []string
	whoAmI             func() ([]byte, error)
	rotateUserAPIKey   func(t *testing.T, userID string) ([]byte, error)
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
		rotateUserAPIKey: func(t *testing.T, userID string) ([]byte, error) {
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
		rotateUserAPIKey: func(t *testing.T, userID string) ([]byte, error) {
			// Assert on arguments
			assert.Equal(t, "dev-user", userID)

			return []byte("test-api-key"), nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "test-api-key")
		},
	},
	// BEGIN COMPATIBILITY WITH PYTHON CLI
	{
		name: "rotate API key for specified user with -i flag",
		args: []string{"user", "rotate-api-key", "-i", "dev-user"},
		rotateUserAPIKey: func(t *testing.T, userID string) ([]byte, error) {
			// Assert on arguments
			assert.Equal(t, "dev-user", userID)

			return []byte("test-api-key"), nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "test-api-key")
			assert.Contains(t, stdout, "test-api-key")
			assert.Contains(t, stdout, "deprecated")
			assert.Contains(t, stdout, "Use -u / --user-id instead")
		},
	},
	{
		name: "rotate API key for specified user with --id flag",
		args: []string{"user", "rotate-api-key", "--id=dev-user"},
		rotateUserAPIKey: func(t *testing.T, userID string) ([]byte, error) {
			// Assert on arguments
			assert.Equal(t, "dev-user", userID)

			return []byte("test-api-key"), nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "test-api-key")
			assert.Contains(t, stdout, "deprecated")
			assert.Contains(t, stdout, "Use -u / --user-id instead")
		},
	},
	// END COMPATIBILITY WITH PYTHON CLI
	{
		name: "client error when rotating API key",
		args: []string{"user", "rotate-api-key", "--user-id=dev-user"},
		rotateUserAPIKey: func(t *testing.T, userID string) ([]byte, error) {
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
				whoAmI:           tc.whoAmI,
				rotateUserAPIKey: tc.rotateUserAPIKey,
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

var userChangePasswordCmdTestCases = []struct {
	name                      string
	args                      []string
	promptResponses           []promptResponse
	changeCurrentUserPassword func(*testing.T, string) ([]byte, error)
	clientFactoryError        error
	assert                    func(t *testing.T, stdout, stderr string, err error)
}{
	{
		name: "display help",
		args: []string{"user", "change-password", "--help"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "change password for logged in user",
		args: []string{"user", "change-password", "-p", "SUp3r$3cr3t!!"},
		changeCurrentUserPassword: func(t *testing.T, newPassword string) ([]byte, error) {
			// Assert on arguments
			assert.Equal(t, "SUp3r$3cr3t!!", newPassword)

			return []byte(""), nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "Password changed")
		},
	},
	{
		name: "change password using prompt",
		args: []string{"user", "change-password"},
		promptResponses: []promptResponse{
			{
				prompt:   "Please enter a new password (it will not be echoed)",
				response: "SUp3r$3cr3t!!",
			},
		},
		changeCurrentUserPassword: func(t *testing.T, newPassword string) ([]byte, error) {
			// Assert on arguments
			assert.Equal(t, "SUp3r$3cr3t!!", newPassword)

			return []byte(""), nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "Password changed")
		},
	},
	{
		name: "client error",
		args: []string{"user", "change-password", "-p", "SUp3r$3cr3t!!"},
		changeCurrentUserPassword: func(t *testing.T, newPassword string) ([]byte, error) {
			return nil, fmt.Errorf("%s", "an error")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: an error\n")
		},
	},
	{
		name:               "client factory error",
		args:               []string{"user", "change-password", "-p", "SUp3r$3cr3t!!"},
		clientFactoryError: fmt.Errorf("%s", "client factory error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: client factory error\n")
		},
	},
}

func TestUserChangePasswordCmd(t *testing.T) {
	for _, tc := range userChangePasswordCmdTestCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := mockUserClient{
				t:                         t,
				changeCurrentUserPassword: tc.changeCurrentUserPassword,
			}

			cmd := newUserCmd(
				func(cmd *cobra.Command) (userClient, error) {
					return mockClient, tc.clientFactoryError
				},
			)

			stdout, stderr, err := executeCommandForTestWithPromptResponses(
				t, cmd, tc.promptResponses, tc.args...,
			)

			tc.assert(t, stdout, stderr, err)
		})
	}
}
