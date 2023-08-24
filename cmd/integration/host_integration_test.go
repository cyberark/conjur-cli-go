//go:build integration
// +build integration

package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHostIntegration(t *testing.T) {
	cli := newConjurTestCLI(t)
	cli.InitAndLoginAsAdmin(t)


	t.Run("rotate own api key", func(t *testing.T) {
		cli.LoginAsHost(t, "bob")

		stdOut, stdErr, err := cli.Run("whoami")
		assert.Contains(t, stdOut, `"username": "host/bob"`)

		priorApiKey := ""
		stdOut, stdErr, err = cli.Run("host", "rotate-api-key")
		assertAPIKeyRotationCmd(t, err, stdOut, stdErr, priorApiKey)
	})

	t.Run("rotate host api key", func(t *testing.T) {
		cli.LoginAsAdmin(t)

		stdOut, stdErr, err := cli.Run("whoami")
		assert.Contains(t, stdOut, `"username": "admin"`)

		priorAPIKey := ""
		stdOut, stdErr, err = cli.Run("host", "rotate-api-key", "-i", fmt.Sprintf("%s:host:bob", cli.account))
		assertAPIKeyRotationCmd(t, err, stdOut, stdErr, priorAPIKey)

		priorAPIKey = stdOut
		stdOut, stdErr, err = cli.Run("host", "rotate-api-key", "-i", "host:bob")
		assertAPIKeyRotationCmd(t, err, stdOut, stdErr, priorAPIKey)

		priorAPIKey = stdOut
		stdOut, stdErr, err = cli.Run("host", "rotate-api-key", "-i", "bob")
		assertAPIKeyRotationCmd(t, err, stdOut, stdErr, priorAPIKey)
	})
}
