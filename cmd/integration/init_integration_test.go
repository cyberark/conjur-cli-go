//go:build integration
// +build integration

package main

import (
	"github.com/cyberark/conjur-api-go/conjurapi"
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
		stdOut, stdErr, err := cli.Run("init", string(conjurapi.EnvironmentCE), "-a", cli.account, "-u", "https://proxy", "--force-netrc", "--force")
		assert.Error(t, err)
		assert.Contains(t, stdOut, "Trust this certificate? [y/N]")

		stdOut, stdErr, err = cli.Run("init", string(conjurapi.EnvironmentCE), "-a", cli.account, "-u", "https://proxy", "--force-netrc", "--force", "--self-signed")
		assert.NotContains(t, stdErr, "Unable to retrieve and validate certificate")
		assert.NotContains(t, stdOut, "Trust this certificate? [y/N]")
		assert.Contains(t, stdOut, "Wrote certificate to ")
		assert.Contains(t, stdErr, selfSignedWarning)
	})

	t.Run("init with insecure flag", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("init", string(conjurapi.EnvironmentCE), "-a", cli.account, "-u", "http://conjur", "-i", "--force-netrc", "--force")
		assertInitCmd(t, err, stdOut, cli.homeDir)
		assert.Equal(t, insecureModeWarning, stdErr)
	})
}
