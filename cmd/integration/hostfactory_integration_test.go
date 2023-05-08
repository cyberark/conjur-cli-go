//go:build integration
// +build integration

package main

import (
	"encoding/json"
	"testing"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/stretchr/testify/assert"
)

func TestHostfactoryIntegration(t *testing.T) {
	cli := newConjurTestCLI(t)
	cli.InitAndLoginAsAdmin(t)
	cli.LoadPolicy(t, `
- !layer test-layer
- !host-factory
  id: test-host-factory
  layers: [ !layer test-layer ]
`)

	t.Run("create token and host", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("hostfactory", "tokens", "create", "--hostfactory-id", "test-host-factory", "--duration", "5m")
		assert.NoError(t, err)
		assert.Empty(t, stdErr)
		assert.Contains(t, stdOut, "expiration")
		assert.Contains(t, stdOut, "cidr")
		assert.Contains(t, stdOut, "token")

		var tokenResponse []conjurapi.HostFactoryTokenResponse
		err = json.Unmarshal([]byte(stdOut), &tokenResponse)
		assert.NoError(t, err)
		token := tokenResponse[0].Token

		stdOut, stdErr, err = cli.Run("hostfactory", "hosts", "create", "--id", "testHost", "--token", token)
		assert.NoError(t, err)
		assert.Empty(t, stdErr)
		assert.Contains(t, stdOut, "created_at")
		assert.Contains(t, stdOut, "owner")
		assert.Contains(t, stdOut, "permissions")
		assert.Contains(t, stdOut, "annotations")
		assert.Contains(t, stdOut, "restricted_to")
		assert.Contains(t, stdOut, "api_key")

		var hostResponse conjurapi.HostFactoryHostResponse
		err = json.Unmarshal([]byte(stdOut), &hostResponse)
		assert.NoError(t, err)
		apiKey := hostResponse.ApiKey

		// Test that the host can authenticate with the API key
		stdOut, stdErr, err = cli.Run("login", "-i", "host/testHost", "-p", apiKey)
		assertLoginCmd(t, err, stdOut, stdErr)

		stdOut, stdErr, err = cli.Run("whoami")
		assertWhoamiCmd(t, err, stdOut, stdErr)
		assert.Contains(t, stdOut, "host/testHost")

		// Log back in as admin
		cli.LoginAsAdmin(t)

		// Revoke the token
		stdOut, stdErr, err = cli.Run("hostfactory", "tokens", "revoke", "--token", token)
		assert.NoError(t, err)
		assert.Empty(t, stdErr)
		assert.Equal(t, stdOut, "Token has been revoked.\n")
	})
}
