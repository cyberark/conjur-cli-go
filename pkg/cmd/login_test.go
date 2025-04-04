package cmd

import (
	"fmt"
	"testing"
	"time"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-api-go/conjurapi/authn"
	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/stretchr/testify/assert"
)

type mockLoginClient struct {
	t                       *testing.T
	loginWithPromptFallback func(t *testing.T, client clients.ConjurClient, username string, password string) (*authn.LoginPair, error)
	oidcLogin               func(t *testing.T, client clients.ConjurClient, username string, password string) (clients.ConjurClient, error)
	jwtAuthenticate         func(t *testing.T, client clients.ConjurClient) error
}

func (m mockLoginClient) LoginWithPromptFallback(client clients.ConjurClient, username string, password string) (*authn.LoginPair, error) {
	return m.loginWithPromptFallback(m.t, client, username, password)
}

func (m mockLoginClient) OidcLogin(client clients.ConjurClient, username string, password string) (clients.ConjurClient, error) {
	return m.oidcLogin(m.t, client, username, password)
}

func (m mockLoginClient) JWTAuthenticate(client clients.ConjurClient) error {
	return m.jwtAuthenticate(m.t, client)
}

var defaultConjurConfig = conjurapi.Config{
	Account:      "dev",
	ApplianceURL: "https://conjur",
}

var oidcConjurConfig = conjurapi.Config{
	Account:      "dev",
	ApplianceURL: "https://conjur",
	AuthnType:    "oidc",
	ServiceID:    "test-service",
}

var jwtConjurConfig = conjurapi.Config{
	Account:      "dev",
	ApplianceURL: "https://conjur",
	AuthnType:    "jwt",
	ServiceID:    "test-service",
	JWTFilePath:  "jwt-file",
}

var loginTestCases = []struct {
	name                    string
	args                    []string
	conjurConfig            conjurapi.Config
	oidcLogin               func(t *testing.T, client clients.ConjurClient, username string, password string) (clients.ConjurClient, error)
	jwtAuthenticate         func(t *testing.T, client clients.ConjurClient) error
	loginWithPromptFallback func(t *testing.T, client clients.ConjurClient, username string, password string) (*authn.LoginPair, error)
	assert                  func(t *testing.T, stdout string, stderr string, err error)
}{
	{
		name: "login command help",
		args: []string{"login", "--help"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name:         "login",
		args:         []string{"login", "-i", "alice", "-p", "secret"},
		conjurConfig: defaultConjurConfig,
		loginWithPromptFallback: func(t *testing.T, client clients.ConjurClient, username string, password string) (*authn.LoginPair, error) {
			// Assert on arguments
			assert.Equal(t, "alice", username)
			assert.Equal(t, "secret", password)

			return &authn.LoginPair{}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.NoError(t, err)
			assert.Empty(t, stderr)
			assert.Contains(t, stdout, "Logged in")
		},
	},
	{
		name:         "login returns error",
		args:         []string{"login", "-i", "alice", "-p", "secret"},
		conjurConfig: defaultConjurConfig,
		loginWithPromptFallback: func(t *testing.T, client clients.ConjurClient, username string, password string) (*authn.LoginPair, error) {
			return nil, assert.AnError
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.EqualError(t, err, assert.AnError.Error())
		},
	},
	{
		name:         "login with debug flag",
		args:         []string{"--debug", "login", "-i", "alice", "-p", "secret"},
		conjurConfig: defaultConjurConfig,
		loginWithPromptFallback: func(t *testing.T, client clients.ConjurClient, username string, password string) (*authn.LoginPair, error) {
			// Perform the login request which should cause the HTTP request and response to be printed
			client.Login(username, password)
			return &authn.LoginPair{}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.NoError(t, err)
			// Stderr should contain the debug output which includes the HTTP request and response
			assert.Contains(t, stderr, "GET /authn/dev/login")
			assert.Contains(t, stdout, "Logged in")
		},
	},
	{
		name:         "login with oidc",
		args:         []string{"login", "-i", "alice", "-p", "secret"},
		conjurConfig: oidcConjurConfig,
		oidcLogin: func(t *testing.T, client clients.ConjurClient, username string, password string) (clients.ConjurClient, error) {
			// Assert on arguments
			assert.Equal(t, "alice", username)
			assert.Equal(t, "secret", password)

			return client, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.NoError(t, err)
			assert.Empty(t, stderr)
			assert.Contains(t, stdout, "Logged in")
		},
	},
	{
		name:         "login with jwt",
		args:         []string{"login"},
		conjurConfig: jwtConjurConfig,
		jwtAuthenticate: func(t *testing.T, client clients.ConjurClient) error {
			// Just return nil to simulate successful JWT authentication
			return nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.NoError(t, err)
			assert.Empty(t, stderr)
			assert.Contains(t, stdout, "Logged in")
		},
	},
	{
		name:         "login with jwt fails",
		args:         []string{"login"},
		conjurConfig: jwtConjurConfig,
		jwtAuthenticate: func(t *testing.T, client clients.ConjurClient) error {
			return fmt.Errorf("jwt authentication failed")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Empty(t, stdout)
			assert.Contains(t, stderr, "Unable to authenticate with Conjur using the provided JWT file: jwt authentication failed")
		},
	},
}

func TestLoginCmd(t *testing.T) {
	t.Parallel()

	for _, tc := range loginTestCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := mockLoginClient{
				t:                       t,
				loginWithPromptFallback: tc.loginWithPromptFallback,
				oidcLogin:               tc.oidcLogin,
				jwtAuthenticate:         tc.jwtAuthenticate,
			}

			cmd := newLoginCmd(
				loginCmdFuncs{
					LoginWithPromptFallback: mockClient.LoginWithPromptFallback,
					OidcLogin:               mockClient.OidcLogin,
					JWTAuthenticate:         mockClient.JWTAuthenticate,
					LoadAndValidateConjurConfig: func(time.Duration) (conjurapi.Config, error) {
						return tc.conjurConfig, nil
					},
				},
			)

			stdout, stderr, err := executeCommandForTest(t, cmd, tc.args...)
			tc.assert(t, stdout, stderr, err)
		})
	}
}
