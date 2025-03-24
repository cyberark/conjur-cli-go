//go:build integration
// +build integration

package main

import (
	"testing"
)

func TestLogoutIntegration(t *testing.T) {
	cli := newConjurTestCLI(t)
	cli.InitAndLoginAsAdmin(t)

	t.Run("logout", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("logout")
		assertLogoutCmd(t, err, stdOut, stdErr)
		assertAuthTokenPurged(t, err, cli.homeDir)
	})
}
