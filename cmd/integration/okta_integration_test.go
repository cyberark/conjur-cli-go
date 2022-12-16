//go:build integration
// +build integration

package main

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegrationOkta(t *testing.T) {
	var (
		stdOut string
		stdErr string
		err    error
	)
	tmpDir := t.TempDir()

	conjurCLI := newConjurCLI(tmpDir)

	t.Run("ensure binary exists", func(t *testing.T) {
		_, err = exec.LookPath(pathToBinary)
		assert.NoError(t, err)
	})

	t.Run("whoami before init", func(t *testing.T) {
		stdOut, stdErr, err = conjurCLI.Run("whoami")
		assert.Error(t, err)
		assert.Equal(t, "", stdOut)
		assert.Equal(t, "Error: Must specify an ApplianceURL -- Must specify an Account\n", stdErr)
	})

	t.Run("init", func(t *testing.T) {
		stdOut, stdErr, err = conjurCLI.Run("init", "-a", "test", "-u", "http://conjur", "-t", "oidc", "--service-id", "okta-2", "--force")
		assert.NoError(t, err)
		assert.Equal(t, "Wrote configuration to "+tmpDir+"/.conjurrc\n", stdOut)
		assert.Equal(t, "", stdErr)
	})

	t.Run("login", func(t *testing.T) {
		stdOut, stdErr, err = conjurCLI.Run("login", "-u", os.Getenv("OKTA_USERNAME"), "-p", os.Getenv("OKTA_PASSWORD"))

		assert.NoError(t, err)
		assert.Contains(t, stdOut, "Logged in\n")
	})

	t.Run("whoami after login", func(t *testing.T) {
		stdOut, stdErr, err = conjurCLI.Run("whoami")
		assert.NoError(t, err)
		assert.Contains(t, stdOut, "token_issued_at")
		assert.Contains(t, stdOut, "client_ip")
		assert.Contains(t, stdOut, "user_agent")
		assert.Contains(t, stdOut, "account")
		assert.Contains(t, stdOut, "username")
		assert.Equal(t, "", stdErr)
	})
}
