//go:build integration
// +build integration

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthenticatorIntegration(t *testing.T) {
	cli := newConjurTestCLI(t)
	cli.InitAndLoginAsAdmin(t)

	t.Run("enable an authenticator", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("authenticator", "enable", "-i", "authn-iam/prod")
		assert.NoError(t, err)
		assert.Empty(t, stdErr)
		assert.Contains(t, stdOut, "Authenticator authn-iam/prod is enabled")
	})

	t.Run("disable an authenticator", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("authenticator", "disable", "-i", "authn-iam/prod")
		assert.NoError(t, err)
		assert.Empty(t, stdErr)
		assert.Contains(t, stdOut, "Authenticator authn-iam/prod is disabled")
	})

	t.Run("enable an authenticator with invalid ID", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("authenticator", "enable", "-i", "invalid-authenticator-id")
		assert.Error(t, err)
		assert.Empty(t, stdOut)
		assert.NotEmpty(t, stdErr)
		assert.Contains(t, stdErr, "invalid authenticator ID format: invalid-authenticator-id, expected format is 'authenticator_type/service_id'\n")
	})

	t.Run("disable an authenticator with invalid ID", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("authenticator", "disable", "-i", "invalid-authenticator-id")
		assert.Error(t, err)
		assert.Empty(t, stdOut)
		assert.NotEmpty(t, stdErr)
		assert.Contains(t, stdErr, "invalid authenticator ID format: invalid-authenticator-id, expected format is 'authenticator_type/service_id'\n")
	})

	t.Run("enable an non-existent authenticator", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("authenticator", "enable", "-i", "authn-iam/non-existent")
		assert.Error(t, err)
		assert.Empty(t, stdOut)
		assert.NotEmpty(t, stdErr)
		assert.Contains(t, stdErr, "401 Unauthorized")
	})

	t.Run("disable an non-existent authenticator", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("authenticator", "disable", "-i", "authn-iam/non-existent")
		assert.Error(t, err)
		assert.Empty(t, stdOut)
		assert.NotEmpty(t, stdErr)
		assert.Contains(t, stdErr, "401 Unauthorized")
	})
}
