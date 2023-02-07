package cmd

import (
	"testing"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-api-go/conjurapi/authn"
	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/cyberark/conjur-cli-go/pkg/prompts"
	"github.com/stretchr/testify/assert"
)

type mockLoginClient struct {
	t                       *testing.T
	loginWithPromptFallback func(t *testing.T, decoratePrompt prompts.DecoratePromptFunc, client clients.ConjurClient, username string, password string) (*authn.LoginPair, error)
	oidcLogin               func(t *testing.T, client clients.ConjurClient, username string, password string) (clients.ConjurClient, error)
}

func (m mockLoginClient) LoginWithPromptFallback(decoratePrompt prompts.DecoratePromptFunc, client clients.ConjurClient, username string, password string) (*authn.LoginPair, error) {
	return m.loginWithPromptFallback(m.t, decoratePrompt, client, username, password)
}

func (m mockLoginClient) OidcLogin(client clients.ConjurClient, username string, password string) (clients.ConjurClient, error) {
	return m.oidcLogin(m.t, client, username, password)
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

var loginTestCases = []struct {
	name                    string
	args                    []string
	conjurConfig            conjurapi.Config
	oidcLogin               func(t *testing.T, client clients.ConjurClient, username string, password string) (clients.ConjurClient, error)
	loginWithPromptFallback func(t *testing.T, decoratePrompt prompts.DecoratePromptFunc, client clients.ConjurClient, username string, password string) (*authn.LoginPair, error)
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
		args:         []string{"login", "-u", "alice", "-p", "secret"},
		conjurConfig: defaultConjurConfig,
		loginWithPromptFallback: func(t *testing.T, decoratePrompt prompts.DecoratePromptFunc, client clients.ConjurClient, username string, password string) (*authn.LoginPair, error) {
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
		args:         []string{"login", "-u", "alice", "-p", "secret"},
		conjurConfig: defaultConjurConfig,
		loginWithPromptFallback: func(t *testing.T, decoratePrompt prompts.DecoratePromptFunc, client clients.ConjurClient, username string, password string) (*authn.LoginPair, error) {
			return nil, assert.AnError
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.EqualError(t, err, assert.AnError.Error())
		},
	},
	{
		name:         "login with debug flag",
		args:         []string{"--debug", "login", "-u", "alice", "-p", "secret"},
		conjurConfig: defaultConjurConfig,
		loginWithPromptFallback: func(t *testing.T, decoratePrompt prompts.DecoratePromptFunc, client clients.ConjurClient, username string, password string) (*authn.LoginPair, error) {
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
	// BEGIN COMPATIBILITY WITH PYTHON CLI
	{
		name:         "login with -i flag",
		args:         []string{"login", "-i", "alice", "-p", "secret"},
		conjurConfig: defaultConjurConfig,
		loginWithPromptFallback: func(t *testing.T, decoratePrompt prompts.DecoratePromptFunc, client clients.ConjurClient, username string, password string) (*authn.LoginPair, error) {
			// Assert on arguments
			assert.Equal(t, "alice", username)
			assert.Equal(t, "secret", password)

			return &authn.LoginPair{}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.NoError(t, err)
			assert.Contains(t, stdout, "deprecated")
			assert.Contains(t, stdout, "Use -u / --username instead")
			assert.Contains(t, stdout, "Logged in")
		},
	},
	{
		name:         "login with --id flag",
		args:         []string{"login", "--id", "alice", "-p", "secret"},
		conjurConfig: defaultConjurConfig,
		loginWithPromptFallback: func(t *testing.T, decoratePrompt prompts.DecoratePromptFunc, client clients.ConjurClient, username string, password string) (*authn.LoginPair, error) {
			// Assert on arguments
			assert.Equal(t, "alice", username)
			assert.Equal(t, "secret", password)

			return &authn.LoginPair{}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.NoError(t, err)
			assert.Contains(t, stdout, "deprecated")
			assert.Contains(t, stdout, "Use -u / --username instead")
			assert.Contains(t, stdout, "Logged in")
		},
	},
	// END COMPATIBILITY WITH PYTHON CLI
	{
		name:         "login with oidc",
		args:         []string{"login", "-u", "alice", "-p", "secret"},
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
}

func TestLoginCmd(t *testing.T) {
	t.Parallel()

	for _, tc := range loginTestCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := mockLoginClient{t: t, loginWithPromptFallback: tc.loginWithPromptFallback, oidcLogin: tc.oidcLogin}

			cmd := newLoginCmd(
				loginCmdFuncs{
					LoginWithPromptFallback: mockClient.LoginWithPromptFallback,
					OidcLogin:               mockClient.OidcLogin,
					LoadAndValidateConjurConfig: func() (conjurapi.Config, error) {
						return tc.conjurConfig, nil
					},
				},
			)

			stdout, stderr, err := executeCommandForTest(t, cmd, tc.args...)
			tc.assert(t, stdout, stderr, err)
		})
	}
}
