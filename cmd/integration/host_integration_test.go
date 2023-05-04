//go:build integration
// +build integration

package main

import (
	"fmt"
	"testing"
)

func TestHostIntegration(t *testing.T) {
	cli := newConjurTestCLI(t)
	cli.InitAndLoginAsAdmin(t)

	t.Run("rotate host api key", func(t *testing.T) {
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
}
