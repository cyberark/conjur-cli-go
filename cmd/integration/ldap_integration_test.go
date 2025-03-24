//go:build integration
// +build integration

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testLDAPLogin(t *testing.T, cli *testConjurCLI) {
	t.Run("init", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run(
			"init", "-a", cli.account,
			"-u", "http://conjur",
			"-t", "ldap",
			"--service-id=test-service",
			"--force-netrc",
			"-i", "--force",
		)
		assertInitCmd(t, err, stdOut, cli.homeDir)
		assert.Equal(t, insecureModeWarning, stdErr)
	})

	t.Run("login", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("login", "-i", "alice", "-p", "aliceLdapPwd")
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

func testLDAPAuthenticatedCli(t *testing.T, cli *testConjurCLI) {
	t.Run("rotate own api key", func(t *testing.T) {
		priorApiKey := ""
		stdOut, stdErr, err := cli.Run("user", "rotate-api-key")
		assertAPIKeyRotationCmd(t, err, stdOut, stdErr, priorApiKey)
	})
}

func testLDAPLogout(t *testing.T, tmpDir string, cli *testConjurCLI) {
	t.Run("logout", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("logout")
		assertLogoutCmd(t, err, stdOut, stdErr)
		assertAuthTokenPurged(t, err, tmpDir)
	})
}

func TestLDAPIntegration(t *testing.T) {
	cli := newConjurTestCLI(t)

	setupLDAPAuthenticator(cli.account)

	testLDAPLogin(t, cli)
	testLDAPAuthenticatedCli(t, cli)
	testLDAPLogout(t, cli.homeDir, cli)
}

func setupLDAPAuthenticator(account string) {
	loadPolicyFile(account, "../../ci/ldap/policy.yml")
}
