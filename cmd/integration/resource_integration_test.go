//go:build integration
// +build integration

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResourceIntegration(t *testing.T) {
	cli := newConjurTestCLI(t)
	cli.InitAndLoginAsAdmin(t)

	t.Run("resource exists returns true", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("resource", "exists", cli.account+":variable:meow")
		assert.NoError(t, err)
		assert.Empty(t, stdErr)
		assert.Equal(t, "true\n", stdOut)
	})

	t.Run("resource exists returns false", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("resource", "exists", cli.account+":variable:meow2")
		assert.NoError(t, err)
		assert.Empty(t, stdErr)
		assert.Equal(t, "false\n", stdOut)
	})

	t.Run("resource exists with json", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("resource", "exists", "--json", cli.account+":variable:meow")
		assert.NoError(t, err)
		assert.Empty(t, stdErr)
		assert.Equal(t, "{\n  \"exists\": true\n}\n", stdOut)
	})

	t.Run("resource permitted roles", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("resource", "permitted-roles", cli.account+":variable:meow", "execute")
		assert.NoError(t, err)
		assert.Empty(t, stdErr)
		assert.Contains(t, stdOut, cli.account+":user:admin")
		assert.NotContains(t, stdOut, cli.account+":user:alice")
	})

	t.Run("resource show", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("resource", "show", cli.account+":variable:meow")
		assert.NoError(t, err)
		assert.Empty(t, stdErr)
		assert.Contains(t, stdOut, `"id": "`+cli.account+`:variable:meow"`)
		assertContainsMetadata(t, stdOut, cli)
	})
}
