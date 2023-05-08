//go:build integration
// +build integration

package main

import (
	"testing"
)

func TestWhoamiIntegration(t *testing.T) {
	cli := newConjurTestCLI(t)
	cli.InitAndLoginAsAdmin(t)

	t.Run("whoami", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("whoami")
		assertWhoamiCmd(t, err, stdOut, stdErr)
	})
}
