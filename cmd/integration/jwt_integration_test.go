//go:build integration
// +build integration

package main

import (
	"fmt"
	"github.com/cyberark/conjur-api-go/conjurapi"
	"testing"

	"github.com/stretchr/testify/assert"
)

type jwtConnection struct {
	jwksURI          string
	tokenAppProperty string
	issuer           string
}

type authnJwtConfig struct {
	hostID      string
	jwtFilePath string
}

func testJwtLogin(t *testing.T, cli *testConjurCLI, jwtConfig authnJwtConfig) {
	t.Run("init", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run(
			"init", string(conjurapi.EnvironmentSH), "-a", cli.account,
			"-u", "http://conjur",
			"-t", "jwt",
			"--jwt-file", jwtConfig.jwtFilePath,
			"--service-id=keycloak",
			"--force-netrc",
			"-i", "--force",
		)
		assertInitCmd(t, err, stdOut, cli.homeDir)
		assert.Contains(t, stdErr, insecureModeWarning)
	})

	t.Run("login", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("login")
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

func testJwtAuthenticatedCli(t *testing.T, cli *testConjurCLI, jwtConfig authnJwtConfig) {
	t.Run("get variable before policy load", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("variable", "get", "-i", "meow")
		assertNotFound(t, err, stdOut, stdErr)
	})

	t.Run("policy load", func(t *testing.T) {
		variablePolicy := fmt.Sprintf(`
- !variable meow

- !permit
  role: !host %s
  privilege: [ read, execute, update ]
  resource: !variable meow`, jwtConfig.hostID)
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

func testJwtLogout(t *testing.T, tmpDir string, cli *testConjurCLI) {
	t.Run("logout", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("logout")
		assertLogoutCmd(t, err, stdOut, stdErr)
		assertAuthTokenPurged(t, err, tmpDir)
	})
}

func TestJWTIntegration(t *testing.T) {
	TestCases := []struct {
		description   string
		jwtConnection jwtConnection
		jwtConfig     authnJwtConfig
		envVars       []string
	}{
		{
			description: "conjur cli user authenticates with jwt",
			jwtConnection: jwtConnection{
				jwksURI:          "http://keycloak:8080/auth/realms/master/protocol/openid-connect/certs",
				tokenAppProperty: "preferred_username",
				issuer:           "http://keycloak:8080/auth/realms/master",
			},
			jwtConfig: authnJwtConfig{
				jwtFilePath: "/jwt",
				hostID:      "service-account-conjurclient",
			},
			envVars: []string{},
		},
	}

	for _, tc := range TestCases {
		t.Run(tc.description, func(t *testing.T) {
			cli := newConjurTestCLI(t)

			setupJwtAuthenticator(cli.account, tc.jwtConnection)

			testJwtLogin(t, cli, tc.jwtConfig)
			testJwtAuthenticatedCli(t, cli, tc.jwtConfig)
			testJwtLogout(t, cli.homeDir, cli)
		})
	}
}

func setupJwtAuthenticator(account string, jwtConnection jwtConnection) {
	loadPolicyFile(account, "../../ci/jwt/policy.yml")

	createSecret(account, "conjur/authn-jwt/keycloak/jwks-uri", jwtConnection.jwksURI)
	createSecret(account, "conjur/authn-jwt/keycloak/token-app-property", jwtConnection.tokenAppProperty)
	createSecret(account, "conjur/authn-jwt/keycloak/issuer", jwtConnection.issuer)
}
