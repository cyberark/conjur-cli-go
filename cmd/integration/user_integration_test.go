//go:build integration
// +build integration

package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserIntegration(t *testing.T) {
	cli := newConjurTestCLI(t)
	cli.InitAndLoginAsAdmin(t)

	t.Run("rotate own api key", func(t *testing.T) {
		priorApiKey := ""
		stdOut, stdErr, err := cli.Run("user", "rotate-api-key")
		assertAPIKeyRotationCmd(t, err, stdOut, stdErr, priorApiKey)
	})

	t.Run("rotate another user's api key", func(t *testing.T) {
		// We must be logged in as admin here
		cli.LoginAsAdmin(t)

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

	t.Run("change password", func(t *testing.T) {
		// Log in as non-admin user to change password
		stdOut, stdErr, err := cli.Run("login", "-i", "alice", "-p", makeDevRequest("retrieve_api_key", map[string]string{"role_id": cli.account + ":user:alice"}))
		assertLoginCmd(t, err, stdOut, stdErr)

		stdOut, stdErr, err = cli.Run("user", "change-password", "-p", "Sup3rS3cr3tc0njuR!t3stp@ssw0rd")
		assert.NoError(t, err)
		assert.Empty(t, stdErr)
		assert.Equal(t, "Password changed\n", stdOut)
	})
}
