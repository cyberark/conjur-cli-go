//go:build integration
// +build integration

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckIntegration(t *testing.T) {
	cli := newConjurTestCLI(t)
	cli.InitAndLoginAsAdmin(t)

	t.Run("check permission", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("check", cli.account+":variable:meow", "execute")
		assert.NoError(t, err)
		assert.Empty(t, stdErr)
		assert.Equal(t, "true\n", stdOut)
	})

	t.Run("check permission for role", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("check", "-r", cli.account+":user:alice", cli.account+":variable:meow", "execute")
		assert.NoError(t, err)
		assert.Empty(t, stdErr)
		assert.Equal(t, "false\n", stdOut)
	})
}

func TestCheckIntegrationCloud(t *testing.T) {
	cli := newConjurTestCLI(t)
	cli.InitCloud(t)

	t.Run("check command", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("check")
		assert.Error(t, err)
		assert.Empty(t, stdOut)
		assert.Contains(t, stdErr, "unknown command \"check\" for \"conjur\"\n")
	})
}
