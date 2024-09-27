//go:build integration
// +build integration

package main

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitIntegration(t *testing.T) {
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
		assert.Contains(t, stdOut, "The server's certificate Sha256 fingerprint is")
		assert.Contains(t, stdErr, selfSignedWarning)
	})

	t.Run("init with insecure flag", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("init", "-a", cli.account, "-u", "http://conjur", "-i", "--force-netrc", "--force")
		assertInitCmd(t, err, stdOut, cli.homeDir)
		assert.Equal(t, insecureModeWarning, stdErr)
	})
}
