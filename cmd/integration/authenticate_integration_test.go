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
