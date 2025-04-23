//go:build integration
// +build integration

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompletionIntegrationCloud(t *testing.T) {
	cli := newConjurTestCLI(t)
	cli.InitCloud(t)

	t.Run("completion command", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("completion")
		assert.Error(t, err)
		assert.Empty(t, stdOut)
		assert.Equal(t, "Error: unknown command \"completion\" for \"conjur\"\nRun 'conjur --help' for usage.\n", stdErr)
	})
}
