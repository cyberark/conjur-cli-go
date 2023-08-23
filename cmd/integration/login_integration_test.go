//go:build integration
// +build integration

package main

import (
	"testing"
)

func TestLoginIntegration(t *testing.T) {
	cli := newConjurTestCLI(t)
	cli.Init(t)

	t.Run("login", func(t *testing.T) {
		cli.LoginAsAdmin(t)
	})
}

func TestLoginWithApplianceURLTrailingSlashIntegration(t *testing.T) {
	cli := newConjurTestCLI(t)
	cli.InitWithTrailingSlash(t)

	t.Run("login", func(t *testing.T) {
		cli.LoginAsAdmin(t)
	})
}
