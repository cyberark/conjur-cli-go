//go:build integration
// +build integration

package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegration(t *testing.T) {
	cli := newConjurTestCLI(t)

	t.Run("ensure binary exists", func(t *testing.T) {
		_, err := exec.LookPath(pathToBinary)
		assert.NoError(t, err)
	})

	t.Run("whoami before init", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("whoami")
		assert.Error(t, err)
		assert.Equal(t, "", stdOut)
		assert.Contains(t, stdErr, "Must specify an ApplianceURL")
		assert.Contains(t, stdErr, "Must specify an Account")
	})

	t.Run("init with self-signed cert", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("init", "-a", cli.account, "-u", "https://proxy", "--force-netrc", "--force")
		assert.Error(t, err)
		assert.Equal(t, "", stdOut)
		assert.Contains(t, stdErr, "Unable to retrieve and validate certificate")
		assert.Contains(t, stdErr, "re-run the init command with the `--self-signed` flag")

		stdOut, stdErr, err = cli.Run("init", "-a", cli.account, "-u", "https://proxy", "--force-netrc", "--force", "--self-signed")
		assert.NotContains(t, stdErr, "Unable to retrieve and validate certificate")
		assert.Contains(t, stdOut, "The server's certificate fingerprint is")
		assert.Contains(t, stdErr, selfSignedWarning)
	})

	t.Run("init", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("init", "-a", cli.account, "-u", "http://conjur", "-i", "--force-netrc", "--force")
		assertInitCmd(t, err, stdOut, cli.homeDir)
		assert.Equal(t, insecureModeWarning, stdErr)
	})

	t.Run("login", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("login", "-i", "admin", "-p", makeDevRequest("retrieve_api_key", map[string]string{"role_id": cli.account + ":user:admin"}))
		assertLoginCmd(t, err, stdOut, stdErr)
	})

	t.Run("whoami after login", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("whoami")
		assertWhoamiCmd(t, err, stdOut, stdErr)
	})

	t.Run("authenticate", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("authenticate")
		assertAuthenticateCmd(t, err, stdOut, stdErr)
	})

	t.Run("get variable before policy load", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("variable", "get", "-i", "meow")
		assertNotFound(t, err, stdOut, stdErr)
	})

	t.Run("policy load", func(t *testing.T) {
		stdOut, stdErr, err := cli.RunWithStdin(
			bytes.NewReader([]byte("- !variable meow\n- !variable woof\n- !user alice\n- !host bob")),
			"policy", "load", "-b", "root", "-f", "-",
		)
		assertPolicyLoadCmd(t, err, stdOut, stdErr)
	})

	t.Run("set variable after policy load", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("variable", "set", "-i", "meow", "-v", "moo")
		assertSetVariableCmd(t, err, stdOut, stdErr)
	})

	t.Run("set another variable after policy load", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("variable", "set", "-i", "woof", "-v", "quack")
		assertSetVariableCmd(t, err, stdOut, stdErr)
	})

	t.Run("get variable after policy load", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("variable", "get", "-i", "meow")
		assertGetVariableCmd(t, err, stdOut, stdErr)
	})

	t.Run("get two variables after policy load", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("variable", "get", "-i", "meow,woof")
		assertGetTwoVariablesCmd(t, err, stdOut, stdErr)
	})

	t.Run("exists returns false", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("role", "exists", "dev:user:meow")
		assertExistsCmd(t, err, stdOut, stdErr)
	})

	t.Run("rotate user alice api key", func(t *testing.T) {
		priorAPIKey := ""
		stdOut, stdErr, err := cli.Run("user", "rotate-api-key", "-i", fmt.Sprintf("%s:user:alice", cli.account))
		assertAPIKeyRotationCmd(t, err, stdOut, stdErr, priorAPIKey)

		priorAPIKey = stdOut
		stdOut, stdErr, err = cli.Run("user", "rotate-api-key", "-i", "user:alice")
		assertAPIKeyRotationCmd(t, err, stdOut, stdErr, priorAPIKey)

		priorAPIKey = stdOut
		stdOut, stdErr, err = cli.Run("user", "rotate-api-key", "-i", "alice")
		assertAPIKeyRotationCmd(t, err, stdOut, stdErr, priorAPIKey)
	})

	t.Run("rotate host bob api key", func(t *testing.T) {
		priorAPIKey := ""
		stdOut, stdErr, err := cli.Run("host", "rotate-api-key", "-i", fmt.Sprintf("%s:host:bob", cli.account))
		assertAPIKeyRotationCmd(t, err, stdOut, stdErr, priorAPIKey)

		priorAPIKey = stdOut
		stdOut, stdErr, err = cli.Run("host", "rotate-api-key", "-i", "host:bob")
		assertAPIKeyRotationCmd(t, err, stdOut, stdErr, priorAPIKey)

		priorAPIKey = stdOut
		stdOut, stdErr, err = cli.Run("host", "rotate-api-key", "-i", "bob")
		assertAPIKeyRotationCmd(t, err, stdOut, stdErr, priorAPIKey)
	})

	t.Run("logout", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("logout")
		assertLogoutCmd(t, err, stdOut, stdErr)
	})
}
