//go:build integration
// +build integration

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPubkeysIntegration(t *testing.T) {
	cli := newConjurTestCLI(t)
	cli.InitAndLoginAsAdmin(t)

	t.Run("get public keys", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("pubkeys", "alice")
		assert.NoError(t, err)
		assert.Empty(t, stdErr)
		assert.Equal(t, "\n\n", stdOut)
	})
}

func TestPubkeysIntegrationCloud(t *testing.T) {
	cli := newConjurTestCLI(t)
	cli.InitCloud(t)

	t.Run("pubkeys command", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("pubkeys", "alice")
		assert.Error(t, err)
		assert.Empty(t, stdOut)
		assert.Contains(t, stdErr, "unknown command \"pubkeys\" for \"conjur\"\n")
	})
}
