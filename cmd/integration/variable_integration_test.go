//go:build integration
// +build integration

package main

import (
	"testing"
)

func TestVariableIntegration(t *testing.T) {
	cli := newConjurTestCLI(t)
	cli.InitAndLoginAsAdmin(t)

	t.Run("set variables", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("variable", "set", "-i", "meow", "-v", "moo")
		assertSetVariableCmd(t, err, stdOut, stdErr)

		stdOut, stdErr, err = cli.Run("variable", "set", "-i", "woof", "-v", "quack")
		assertSetVariableCmd(t, err, stdOut, stdErr)
	})

	t.Run("get variable", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("variable", "get", "-i", "meow")
		assertGetVariableCmd(t, err, stdOut, stdErr, "moo")
	})

	t.Run("get two variables", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("variable", "get", "-i", "meow,woof")
		assertGetTwoVariablesCmd(t, err, stdOut, stdErr)
	})

	t.Run("get duplicated variables", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("variable", "get", "-i", "meow,woof,meow")
		assertGetTwoVariablesCmd(t, err, stdOut, stdErr)
	})

	t.Run("get updated variable without version", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("variable", "set", "-i", "meow", "-v", "moo2")
		assertSetVariableCmd(t, err, stdOut, stdErr)

		stdOut, stdErr, err = cli.Run("variable", "get", "-i", "meow")
		assertGetVariableCmd(t, err, stdOut, stdErr, "moo2")
	})

	t.Run("get updated variable with version", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("variable", "get", "-i", "meow", "-v", "1")
		assertGetVariableCmd(t, err, stdOut, stdErr, "moo")

		stdOut, stdErr, err = cli.Run("variable", "get", "-i", "meow", "-v", "2")
		assertGetVariableCmd(t, err, stdOut, stdErr, "moo2")
	})

	t.Run("get variable with nonexistent version", func(t *testing.T) {
		stdOut, stdErr, err := cli.Run("variable", "get", "-i", "meow", "-v", "3")
		assertNotFound(t, err, stdOut, stdErr)
	})

	t.Run("variables with colons", func(t *testing.T) {
		variablePolicy := `
- !variable foo:bar
- !variable foo:bar:baz
`
		cli.LoadPolicy(t, variablePolicy)

		stdOut, stdErr, err := cli.Run("variable", "set", "-i", "foo:bar", "-v", "secretValue")
		assertSetVariableCmd(t, err, stdOut, stdErr)

		stdOut, stdErr, err = cli.Run("variable", "get", "-i", "foo:bar")
		assertGetVariableCmd(t, err, stdOut, stdErr, "secretValue")

		stdOut, stdErr, err = cli.Run("variable", "set", "-i", "foo:bar:baz", "-v", "otherSecretValue")
		assertSetVariableCmd(t, err, stdOut, stdErr)

		stdOut, stdErr, err = cli.Run("variable", "get", "-i", "foo:bar:baz")
		assertGetVariableCmd(t, err, stdOut, stdErr, "otherSecretValue")
	})
}
