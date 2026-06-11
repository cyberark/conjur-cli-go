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
	cloudLogin              func(t *testing.T, client clients.ConjurClient, username string, password string, insecureLogin bool) (clients.ConjurClient, error)
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

func (m mockLoginClient) CloudLogin(client clients.ConjurClient, username string, password string, insecureLogin bool) (clients.ConjurClient, error) {
	return m.cloudLogin(m.t, client, username, password, insecureLogin)
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

var cloudConjurConfig = conjurapi.Config{
	Account:      "conjur",
	ApplianceURL: "https://mytenant.secretsmgr.cyberark.cloud/api",
	AuthnType:    "cloud",
	ServiceID:    "cyberark",
	Environment:  conjurapi.EnvironmentSaaS,
}

var loginTestCases = []struct {
	name                    string
	args                    []string
	conjurConfig            conjurapi.Config
	oidcLogin               func(t *testing.T, client clients.ConjurClient, username string, password string) (clients.ConjurClient, error)
	jwtAuthenticate         func(t *testing.T, client clients.ConjurClient) error
	loginWithPromptFallback func(t *testing.T, client clients.ConjurClient, username string, password string) (*authn.LoginPair, error)
	cloudLogin              func(t *testing.T, client clients.ConjurClient, username string, password string, insecureLogin bool) (clients.ConjurClient, error)
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
			assert.Contains(t, stderr, "Unable to authenticate with Idira Secrets Manager using the provided JWT file: jwt authentication failed")
		},
	},
	{
		name:         "login with cloud",
		args:         []string{"login", "-i", "alice@example.com", "-p", "secret"},
		conjurConfig: cloudConjurConfig,
		cloudLogin: func(t *testing.T, client clients.ConjurClient, username string, password string, insecureLogin bool) (clients.ConjurClient, error) {
			assert.Equal(t, "alice@example.com", username)
			assert.Equal(t, "secret", password)
			assert.False(t, insecureLogin)
			return client, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.NoError(t, err)
			assert.Empty(t, stderr)
			assert.Contains(t, stdout, "Logged in")
		},
	},
	{
		name:         "login with cloud returns error",
		args:         []string{"login", "-i", "alice@example.com", "-p", "secret"},
		conjurConfig: cloudConjurConfig,
		cloudLogin: func(t *testing.T, client clients.ConjurClient, username string, password string, insecureLogin bool) (clients.ConjurClient, error) {
			return nil, fmt.Errorf("cloud login failed")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "cloud login failed")
		},
	},
	{
		name:         "login with cloud and insecure-bypass-idp-pin flag",
		args:         []string{"login", "-i", "alice@example.com", "-p", "secret", "--insecure-bypass-idp-pin"},
		conjurConfig: cloudConjurConfig,
		cloudLogin: func(t *testing.T, client clients.ConjurClient, username string, password string, insecureLogin bool) (clients.ConjurClient, error) {
			assert.True(t, insecureLogin)
			return client, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.NoError(t, err)
			assert.Contains(t, stderr, "Warning: PIN verification was skipped.")
			assert.Contains(t, stdout, "Logged in")
		},
	},
	{
		name:         "insecure-bypass-idp-pin flag not available for non-cloud",
		args:         []string{"login", "--insecure-bypass-idp-pin"},
		conjurConfig: defaultConjurConfig,
		loginWithPromptFallback: func(t *testing.T, client clients.ConjurClient, username string, password string) (*authn.LoginPair, error) {
			return &authn.LoginPair{}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "unknown flag: --insecure-bypass-idp-pin")
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
				cloudLogin:              tc.cloudLogin,
			}

			cmd := newLoginCmd(
				loginCmdFuncs{
					LoginWithPromptFallback: mockClient.LoginWithPromptFallback,
					OidcLogin:               mockClient.OidcLogin,
					JWTAuthenticate:         mockClient.JWTAuthenticate,
					CloudLogin:              mockClient.CloudLogin,
					LoadAndValidateConjurConfig: func(time.Duration) (conjurapi.Config, error) {
						return tc.conjurConfig, nil
					},
				},
				tc.conjurConfig,
			)

			stdout, stderr, err := executeCommandForTest(t, cmd, tc.args...)
			tc.assert(t, stdout, stderr, err)
		})
	}
}
