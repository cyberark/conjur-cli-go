//go:build integration
// +build integration

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoleIntegration(t *testing.T) {
	cli := newConjurTestCLI(t)
	cli.InitAndLoginAsAdmin(t)
	cli.LoadPolicy(t, `
- !layer test-layer

- !grant
  role: !layer test-layer
  members:
  - !user alice
`)

	t.Run("role exists returns true", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("role", "exists", cli.account+":host:bob")
		assert.NoError(t, err)
		assert.Empty(t, stdErr)
		assert.Equal(t, "true\n", stdOut)
	})

	t.Run("role exists returns false", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("role", "exists", cli.account+":host:bob2")
		assert.NoError(t, err)
		assert.Empty(t, stdErr)
		assert.Equal(t, "false\n", stdOut)
	})

	t.Run("role exists with json", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("role", "exists", "--json", cli.account+":host:bob")
		assert.NoError(t, err)
		assert.Empty(t, stdErr)
		assert.Equal(t, "{\n  \"exists\": true\n}\n", stdOut)
	})

	t.Run("role show", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("role", "show", cli.account+":host:bob")
		assert.NoError(t, err)
		assert.Empty(t, stdErr)
		assert.Contains(t, stdOut, `"role": "`+cli.account+`:host:bob"`)
		assert.Contains(t, stdOut, `"policy": "`+cli.account+`:policy:root"`)
		assert.Contains(t, stdOut, `"member":`)
		assert.Contains(t, stdOut, `"ownership":`)
	})

	t.Run("role members", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("role", "members", cli.account+":layer:test-layer")
		assert.NoError(t, err)
		assert.Empty(t, stdErr)
		assert.Contains(t, stdOut, cli.account+":user:alice")
		assert.NotContains(t, stdOut, "admin_option")
		assert.NotContains(t, stdOut, "role")
		assert.NotContains(t, stdOut, "member")
	})

	t.Run("role members with verbose", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("role", "members", cli.account+":layer:test-layer", "--verbose")
		assert.NoError(t, err)
		assert.Empty(t, stdErr)
		assert.Contains(t, stdOut, cli.account+":user:alice")
		assert.Contains(t, stdOut, "admin_option")
		assert.Contains(t, stdOut, "role")
		assert.Contains(t, stdOut, "member")
	})

	t.Run("role memberships", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("role", "memberships", cli.account+":user:alice")
		assert.NoError(t, err)
		assert.Empty(t, stdErr)
		assert.Contains(t, stdOut, cli.account+":layer:test-layer")
	})
}
