//go:build integration
// +build integration

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWhoamiIntegration(t *testing.T) {
	cli := newConjurTestCLI(t)
	cli.InitAndLoginAsAdmin(t)

	t.Run("whoami", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("whoami")
		assertWhoamiCmd(t, err, stdOut, stdErr)
	})
}

func assertWhoamiCmd(t *testing.T, err error, stdOut string, stdErr string) {
	assert.NoError(t, err)
	assert.Contains(t, stdOut, "token_issued_at")
	assert.Contains(t, stdOut, "client_ip")
	assert.Contains(t, stdOut, "user_agent")
	assert.Contains(t, stdOut, "account")
	assert.Contains(t, stdOut, "username")
	assert.Equal(t, "", stdErr)
}
