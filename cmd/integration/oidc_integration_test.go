//go:build integration
// +build integration

package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

type oidcConnection struct {
	providerURI  string
	clientID     string
	clientSecret string
}

type oidcCredentials struct {
	username string
	password string
}

type authnOidcConfig struct {
	serviceID    string
	claimMapping string
	policyUser   string
}

func testLogin(t *testing.T, cli *testConjurCLI, oc oidcCredentials, aoc authnOidcConfig) {
	t.Run("init", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run(
			"init", "-a", cli.account,
			"-u", "http://conjur",
			"-t", "oidc",
			"--service-id", aoc.serviceID,
			"--force-netrc",
			"-i", "--force",
		)
		assertInitCmd(t, err, stdOut, cli.homeDir)
		assert.Equal(t, insecureModeWarning, stdErr)
	})

	t.Run("login", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("login", "-i", oc.username, "-p", oc.password)
		assertLoginCmd(t, err, stdOut, stdErr)
		assertAuthTokenCached(t, cli.homeDir)
	})

	t.Run("whoami after login", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("whoami")
		assertWhoamiCmd(t, err, stdOut, stdErr)
	})

	t.Run("authenticate", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("authenticate")
		assertAuthenticateCmd(t, err, stdOut, stdErr)
	})
}

func testAuthenticatedCli(t *testing.T, cli *testConjurCLI, aoc authnOidcConfig) {
	t.Run("get variable before policy load", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("variable", "get", "-i", "meow")
		assertNotFound(t, err, stdOut, stdErr)
	})

	t.Run("policy load", func(t *testing.T) {
		variablePolicy := fmt.Sprintf(`
- !variable meow

- !permit
  role: !user %s
  privilege: [ read, execute, update ]
  resource: !variable meow`, aoc.policyUser)
		cli.LoadPolicy(t, variablePolicy)
	})

	t.Run("set variable after policy load", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("variable", "set", "-i", "meow", "-v", "moo")
		assertSetVariableCmd(t, err, stdOut, stdErr)
	})

	t.Run("get variable after policy load", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("variable", "get", "-i", "meow")
		assertGetVariableCmd(t, err, stdOut, stdErr, "moo")
	})
}

func testLogout(t *testing.T, tmpDir string, cli *testConjurCLI, aoc authnOidcConfig) {
	if aoc.serviceID == "keycloak" {
		t.Run("login as another user", func(t *testing.T) {
			// Get the modifieddate of the netrc file
			info, err := os.Stat(tmpDir + "/.netrc")
			assert.NoError(t, err)
			modifiedDate := info.ModTime()

			stdOut, stdErr, err := cli.Run("login", "-i", "bob.somebody", "-p", "bob")
			assertLoginCmd(t, err, stdOut, stdErr)

			// Check that the token file is modified
			info, err = os.Stat(tmpDir + "/.netrc")
			assert.NoError(t, err)
			assert.NotEqual(t, modifiedDate, info.ModTime())
		})

		t.Run("whoami after login as another user", func(t *testing.T) {
			stdOut, stdErr, err := cli.Run("whoami")
			assertWhoamiCmd(t, err, stdOut, stdErr)

			assert.Contains(t, stdOut, "bob")
			assert.NotContains(t, stdOut, "alice")
		})
	}

	t.Run("failed login doesn't modify netrc file", func(t *testing.T) {
		// Get the modifieddate of the netrc file
		info, err := os.Stat(tmpDir + "/.netrc")
		assert.NoError(t, err)
		modifiedDate := info.ModTime()

		_, stdErr, err := cli.Run("login", "-i", "not_in_conjur", "-p", "not_in_conjur")
		assert.Error(t, err)
		assert.NotEmpty(t, stdErr)

		// Check that the netrc file is not modified
		info, err = os.Stat(tmpDir + "/.netrc")
		assert.NoError(t, err)
		assert.Equal(t, modifiedDate, info.ModTime())
	})

	t.Run("logout", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("logout")
		assertLogoutCmd(t, err, stdOut, stdErr)
		assertAuthTokenPurged(t, err, tmpDir)
	})

	t.Run("whoami after logout", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("whoami")
		assert.Error(t, err)
		assert.Contains(t, stdErr, "Please login again")
		assert.Equal(t, "", stdOut)
	})
}

func TestOIDCIntegration(t *testing.T) {
	TestCases := []struct {
		description     string
		oidcConnection  oidcConnection
		oidcCredentials oidcCredentials
		authnOidcConfig authnOidcConfig
		envVars         []string
		beforeFunc      func() error
	}{
		{
			description: "conjur cli user authenticates with keycloak",
			oidcConnection: oidcConnection{
				providerURI:  "https://keycloak:8443/auth/realms/master",
				clientID:     "conjurClient",
				clientSecret: "1234",
			},
			oidcCredentials: oidcCredentials{
				username: "alice",
				password: "alice",
			},
			authnOidcConfig: authnOidcConfig{
				serviceID:    "keycloak",
				claimMapping: "email",
				policyUser:   "alice@conjur.net",
			},
			envVars: []string{},
		},
		{
			description: "conjur cli user authenticates with okta",
			oidcConnection: oidcConnection{
				providerURI:  os.Getenv("OKTA_PROVIDER_URI") + "oauth2/default",
				clientID:     os.Getenv("OKTA_CLIENT_ID"),
				clientSecret: os.Getenv("OKTA_CLIENT_SECRET"),
			},
			oidcCredentials: oidcCredentials{
				username: os.Getenv("OKTA_USERNAME"),
				password: os.Getenv("OKTA_PASSWORD"),
			},
			authnOidcConfig: authnOidcConfig{
				serviceID:    "okta",
				claimMapping: "preferred_username",
				policyUser:   os.Getenv("OKTA_USERNAME"),
			},
			envVars: []string{
				"OKTA_PROVIDER_URI",
				"OKTA_CLIENT_ID",
				"OKTA_CLIENT_SECRET",
				"OKTA_USERNAME",
				"OKTA_PASSWORD",
			},
		},
		{
			description: "conjur cli user authenticates with identity",
			oidcConnection: oidcConnection{
				providerURI:  os.Getenv("IDENTITY_PROVIDER_URI"),
				clientID:     os.Getenv("IDENTITY_CLIENT_ID"),
				clientSecret: os.Getenv("IDENTITY_CLIENT_SECRET"),
			},
			oidcCredentials: oidcCredentials{
				username: os.Getenv("IDENTITY_USERNAME"),
				password: os.Getenv("IDENTITY_PASSWORD"),
			},
			authnOidcConfig: authnOidcConfig{
				serviceID:    "identity",
				claimMapping: "email",
				policyUser:   os.Getenv("IDENTITY_USERNAME"),
			},
			envVars: []string{
				"IDENTITY_PROVIDER_URI",
				"IDENTITY_CLIENT_ID",
				"IDENTITY_CLIENT_SECRET",
				"IDENTITY_USERNAME",
				"IDENTITY_PASSWORD",
			},
			beforeFunc: func() error {
				tmp, err := template.ParseFiles("../../ci/identity/users.template.yml")
				if err != nil {
					return err
				}

				err = os.Remove("../../ci/identity/users.yml")
				if err != nil {
					return err
				}

				file, err := os.Create("../../ci/identity/users.yml")
				if err != nil {
					return err
				}

				defer file.Close()

				err = tmp.Execute(file, map[string]string{"IDENTITY_USERNAME": os.Getenv("IDENTITY_USERNAME")})
				if err != nil {
					return err
				}

				return nil
			},
		},
	}

	for _, tc := range TestCases {
		t.Run(tc.description, func(t *testing.T) {
			cli := newConjurTestCLI(t)

			err := hasValidVariables(tc.envVars)
			assert.Nil(t, err)

			if tc.beforeFunc != nil {
				err := tc.beforeFunc()
				assert.Nil(t, err)
			}

			setupAuthenticator(account, tc.oidcConnection, tc.authnOidcConfig)

			testLogin(t, cli, tc.oidcCredentials, tc.authnOidcConfig)
			testAuthenticatedCli(t, cli, tc.authnOidcConfig)
			testLogout(t, cli.homeDir, cli, tc.authnOidcConfig)
		})
	}
}

func hasValidVariables(keys []string) error {
	empty := []string{}

	for _, key := range keys {
		if os.Getenv(key) == "" {
			empty = append(empty, key)
		}
	}

	if len(empty) > 0 {
		return errors.New("The following environment variables have not been set: " + strings.Join(empty[:], ", "))
	}

	return nil
}

func setupAuthenticator(account string, oc oidcConnection, aoc authnOidcConfig) {
	loadPolicyFile(account, fmt.Sprintf("../../ci/%s/policy.yml", aoc.serviceID))
	loadPolicyFile(account, fmt.Sprintf("../../ci/%s/users.yml", aoc.serviceID))

	createSecret(account, fmt.Sprintf("conjur/authn-oidc/%s/provider-uri", aoc.serviceID), oc.providerURI)
	createSecret(account, fmt.Sprintf("conjur/authn-oidc/%s/client-id", aoc.serviceID), oc.clientID)
	createSecret(account, fmt.Sprintf("conjur/authn-oidc/%s/client-secret", aoc.serviceID), oc.clientSecret)
	createSecret(account, fmt.Sprintf("conjur/authn-oidc/%s/claim-mapping", aoc.serviceID), aoc.claimMapping)
	createSecret(account, fmt.Sprintf("conjur/authn-oidc/%s/redirect_uri", aoc.serviceID), "http://127.0.0.1:8888/callback")
}

func assertAuthTokenCached(t *testing.T, tmpDir string) {
	// Check that the netrc file contains the entry for the conjur server
	contents, err := os.ReadFile(tmpDir + "/.netrc")
	assert.NoError(t, err)
	assert.Contains(t, string(contents), "http://conjur")
}

func assertAuthTokenPurged(t *testing.T, err error, tmpDir string) {
	// Ensure that the netrc file does not contain the entry for the conjur server
	contents, err := os.ReadFile(tmpDir + "/.netrc")
	assert.NoError(t, err)
	assert.NotContains(t, string(contents), "http://conjur")
}
