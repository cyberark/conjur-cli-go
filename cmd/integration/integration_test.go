//go:build integration
// +build integration

package main

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegration(t *testing.T) {
	var (
		stdOut string
		stdErr string
		err    error
	)
	tmpDir := t.TempDir()
	// Lean on the uniqueness of temp directories
	account := strings.Replace(tmpDir, "/", "", -1)

	conjurCLI := newConjurCLI(tmpDir)

	t.Run("ensure binary exists", func(t *testing.T) {
		_, err = exec.LookPath(pathToBinary)
		assert.NoError(t, err)
	})

	cleanUpConjurAccount := prepareConjurAccount(account)
	defer cleanUpConjurAccount()

	t.Run("whoami before init", func(t *testing.T) {
		stdOut, stdErr, err = conjurCLI.Run("whoami")
		assert.Error(t, err)
		assert.Equal(t, "", stdOut)
		assert.Contains(t, stdErr, "Must specify an ApplianceURL")
		assert.Contains(t, stdErr, "Must specify an Account")
	})

	t.Run("init", func(t *testing.T) {
		stdOut, stdErr, err = conjurCLI.Run("init", "-a", account, "-u", "http://conjur", "-i", "--force-netrc", "--force")
		assert.NoError(t, err)
		assert.Equal(t, "Wrote configuration to "+tmpDir+"/.conjurrc\n", stdOut)
		assert.Equal(t, insecureModeWarning, stdErr)
	})

	t.Run("login", func(t *testing.T) {
		stdOut, stdErr, err = conjurCLI.Run("login", "-i", "admin", "-p", makeDevRequest("retrieve_api_key", map[string]string{"role_id": account + ":user:admin"}))
		assertLoginCmd(t, err, stdOut, stdErr)
	})

	t.Run("whoami after login", func(t *testing.T) {
		stdOut, stdErr, err = conjurCLI.Run("whoami")
		assertWhoamiCmd(t, err, stdOut, stdErr)
	})

	t.Run("authenticate", func(t *testing.T) {
		stdOut, stdErr, err = conjurCLI.Run("authenticate")
		assertAuthenticateCmd(t, err, stdOut, stdErr)
	})

	t.Run("get variable before policy load", func(t *testing.T) {
		stdOut, stdErr, err = conjurCLI.Run("variable", "get", "-i", "meow")
		assertNotFound(t, err, stdOut, stdErr)
	})

	t.Run("policy load", func(t *testing.T) {
		stdOut, stdErr, err = conjurCLI.RunWithStdin(
			bytes.NewReader([]byte("- !variable meow")),
			"policy", "load", "-b", "root", "-f", "-",
		)
		assertPolicyLoadCmd(t, err, stdOut, stdErr)
	})

	t.Run("set variable after policy load", func(t *testing.T) {
		stdOut, stdErr, err = conjurCLI.Run("variable", "set", "-i", "meow", "-v", "moo")
		assertSetVariableCmd(t, err, stdOut, stdErr)
	})

	t.Run("get variable after policy load", func(t *testing.T) {
		stdOut, stdErr, err = conjurCLI.Run("variable", "get", "-i", "meow")
		assertGetVariableCmd(t, err, stdOut, stdErr)
	})

	t.Run("exists returns false", func(t *testing.T) {
		stdOut, stdErr, err = conjurCLI.Run("role", "exists", "dev:user:meow")
		assertExistsCmd(t, err, stdOut, stdErr)
	})

	t.Run("logout", func(t *testing.T) {
		stdOut, stdErr, err = conjurCLI.Run("logout")
		assertLogoutCmd(t, err, stdOut, stdErr)
	})
}
