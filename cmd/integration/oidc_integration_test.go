//go:build integration
// +build integration

package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOidcIntegrationKeycloak(t *testing.T) {
	var (
		stdOut string
		stdErr string
		err    error
	)

	tmpDir := t.TempDir()
	account := strings.Replace(tmpDir, "/", "", -1)
	conjurCLI := newConjurCLI(tmpDir)

	cleanUpConjurAccount := prepareConjurAccount(account)
	defer cleanUpConjurAccount()

	// Initialize Keycloak policy/variables
	setupKeycloakAuthenticator(account)

	t.Run("init", func(t *testing.T) {
		stdOut, stdErr, err = conjurCLI.Run(
			"init", "-a", account,
			"-u", "http://conjur",
			"-t", "oidc",
			"--service-id", "keycloak",
			"--force",
		)
		assert.NoError(t, err)
		assert.Equal(t, "Wrote configuration to "+tmpDir+"/.conjurrc\n", stdOut)
		assert.Equal(t, "", stdErr)
	})

	t.Run("login", func(t *testing.T) {
		stdOut, stdErr, err = conjurCLI.Run("login", "-u", "alice", "-p", "alice")
		assertLoginCmd(t, err, stdOut, stdErr)

		// Check that the token file is created with the correct permissions
		info, err := os.Stat(tmpDir + "/.conjur/oidc_token")
		assert.NoError(t, err)
		assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
	})

	t.Run("whoami after login", func(t *testing.T) {
		stdOut, stdErr, err = conjurCLI.Run("whoami")
		assertWhoamiCmd(t, err, stdOut, stdErr)
	})

	t.Run("get variable before policy load", func(t *testing.T) {
		stdOut, stdErr, err = conjurCLI.Run("variable", "get", "-i", "meow")
		assertNotFound(t, err, stdOut, stdErr)
	})

	t.Run("policy load", func(t *testing.T) {
		variablePolicy := `
- !variable meow

- !permit
  role: !user alice@conjur.net
  privilege: [ read, execute, update ]
  resource: !variable meow`
		stdOut, stdErr, err = conjurCLI.RunWithStdin(
			bytes.NewReader([]byte(variablePolicy)),
			"policy", "load", "-b", "root", "-f", "-",
		)
		assertPolicyLoadCmd(t, err, stdOut, stdErr)
	})

	t.Run("set variable after policy load", func(t *testing.T) {
		stdOut, stdErr, err = conjurCLI.Run("variable", "set", "-i", "meow", "-v", "moo")
		assert.NoError(t, err)
		assert.Equal(t, "Value added\n", stdOut)
		assert.Equal(t, "", stdErr)
	})

	t.Run("get variable after policy load", func(t *testing.T) {
		stdOut, stdErr, err = conjurCLI.Run("variable", "get", "-i", "meow")
		assert.NoError(t, err)
		assert.Equal(t, "moo\n", stdOut)
		assert.Equal(t, "", stdErr)
	})

	t.Run("logout", func(t *testing.T) {
		stdOut, stdErr, err = conjurCLI.Run("logout")
		assert.NoError(t, err)
		assert.Contains(t, stdOut, "Logged out\n")

		// Ensure cached token is removed
		_, err = os.Stat(tmpDir + "/.conjur/oidc_token")
		assert.ErrorIs(t, err, os.ErrNotExist)
	})
}

func TestOidcIntegrationOkta(t *testing.T) {
	var (
		stdOut string
		stdErr string
		err    error
	)

	// Skip these tests if the required environment variables are not set
	if os.Getenv("OKTA_PROVIDER_URI") == "" {
		fmt.Println("WARNING: These tests were skipped due to missing Okta variables. Run again with Summon to include these tests.")
		return
	}

	tmpDir := t.TempDir()
	account := strings.Replace(tmpDir, "/", "", -1)
	conjurCLI := newConjurCLI(tmpDir)

	cleanUpConjurAccount := prepareConjurAccount(account)
	defer cleanUpConjurAccount()

	setupOktaAuthenticator(account)

	t.Run("init", func(t *testing.T) {
		stdOut, stdErr, err = conjurCLI.Run(
			"init", "-a", account,
			"-u", "http://conjur",
			"-t", "oidc",
			"--service-id", "okta-2",
			"--force",
		)
		assert.NoError(t, err)
		assert.Equal(t, "Wrote configuration to "+tmpDir+"/.conjurrc\n", stdOut)
		assert.Equal(t, "", stdErr)
	})

	t.Run("login", func(t *testing.T) {
		stdOut, stdErr, err = conjurCLI.Run("login", "-u", os.Getenv("OKTA_USERNAME"), "-p", os.Getenv("OKTA_PASSWORD"))
		assertLoginCmd(t, err, stdOut, stdErr)

		// Check that the token file is created with the correct permissions
		info, err := os.Stat(tmpDir + "/.conjur/oidc_token")
		assert.NoError(t, err)
		assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
	})

	t.Run("whoami after login", func(t *testing.T) {
		stdOut, stdErr, err = conjurCLI.Run("whoami")
		assertWhoamiCmd(t, err, stdOut, stdErr)
	})

	t.Run("get variable before policy load", func(t *testing.T) {
		stdOut, stdErr, err = conjurCLI.Run("variable", "get", "-i", "meow")
		assertNotFound(t, err, stdOut, stdErr)
	})

	t.Run("policy load", func(t *testing.T) {
		variablePolicy := `
- !variable meow

- !permit
  role: !user ` + os.Getenv("OKTA_USERNAME") + `
  privilege: [ read, execute, update ]
  resource: !variable meow`
		stdOut, stdErr, err = conjurCLI.RunWithStdin(
			bytes.NewReader([]byte(variablePolicy)),
			"policy", "load", "-b", "root", "-f", "-",
		)
		assertPolicyLoadCmd(t, err, stdOut, stdErr)
	})

	t.Run("set variable after policy load", func(t *testing.T) {
		stdOut, stdErr, err = conjurCLI.Run("variable", "set", "-i", "meow", "-v", "moo")
		assert.NoError(t, err)
		assert.Equal(t, "Value added\n", stdOut)
		assert.Equal(t, "", stdErr)
	})

	t.Run("get variable after policy load", func(t *testing.T) {
		stdOut, stdErr, err = conjurCLI.Run("variable", "get", "-i", "meow")
		assert.NoError(t, err)
		assert.Equal(t, "moo\n", stdOut)
		assert.Equal(t, "", stdErr)
	})

	t.Run("logout", func(t *testing.T) {
		stdOut, stdErr, err = conjurCLI.Run("logout")
		assert.NoError(t, err)
		assert.Contains(t, stdOut, "Logged out\n")

		// Ensure cached token is removed
		_, err = os.Stat(tmpDir + "/.conjur/oidc_token")
		assert.ErrorIs(t, err, os.ErrNotExist)
	})
}

func setupKeycloakAuthenticator(account string) {
	loadPolicyFile(account, "../../ci/keycloak/policy.yml")
	loadPolicyFile(account, "../../ci/keycloak/users.yml")

	createSecret(account, "conjur/authn-oidc/keycloak/provider-uri", "https://keycloak:8443/auth/realms/master")
	createSecret(account, "conjur/authn-oidc/keycloak/client-id", "conjurClient")
	createSecret(account, "conjur/authn-oidc/keycloak/client-secret", "1234")
	createSecret(account, "conjur/authn-oidc/keycloak/claim-mapping", "email")
	createSecret(account, "conjur/authn-oidc/keycloak/redirect_uri", "http://127.0.0.1:8888/callback")
}

// NOTE: Depends on Summon variables in CLI container
func setupOktaAuthenticator(account string) {
	loadPolicyFile(account, "../../ci/okta/policy.yml")
	loadPolicyFile(account, "../../ci/okta/users.yml")

	createSecret(account, "conjur/authn-oidc/okta-2/provider-uri", os.Getenv("OKTA_PROVIDER_URI")+"oauth2/default")
	createSecret(account, "conjur/authn-oidc/okta-2/client-id", os.Getenv("OKTA_CLIENT_ID"))
	createSecret(account, "conjur/authn-oidc/okta-2/client-secret", os.Getenv("OKTA_CLIENT_SECRET"))
	createSecret(account, "conjur/authn-oidc/okta-2/claim-mapping", "preferred_username")
	createSecret(account, "conjur/authn-oidc/okta-2/nonce", "1656b4264b60af659fce")
	createSecret(account, "conjur/authn-oidc/okta-2/state", "4f413476ef7e2395f0af")
	createSecret(account, "conjur/authn-oidc/okta-2/redirect_uri", "http://127.0.0.1:8888/callback")
}
