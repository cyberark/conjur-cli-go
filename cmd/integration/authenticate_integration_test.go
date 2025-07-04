//go:build integration
// +build integration

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthenticateIntegration(t *testing.T) {
	cli := newConjurTestCLI(t)
	cli.InitAndLoginAsAdmin(t)

	t.Run("authenticate", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("authenticate")
		assertAuthenticateCmd(t, err, stdOut, stdErr)
	})

	t.Run("authenticate with header flag", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("authenticate", "--header")
		assert.NoError(t, err)
		assert.Empty(t, stdErr)
		assert.Regexp(t, `^Authorization: Token token=\"(.+)\"\n$`, stdOut)
	})
}

func TestAuthenticateIntegrationCloud(t *testing.T) {
	cli := newConjurTestCLI(t)
	cli.InitCloud(t)

	t.Run("authenticate command", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("authenticate")
		assert.Error(t, err)
		assert.Empty(t, stdOut)
		assert.Contains(t, stdErr, "unknown command \"authenticate\" for \"conjur\"\n\nDid you mean this?\n\tauthenticator\n")
	})
}
