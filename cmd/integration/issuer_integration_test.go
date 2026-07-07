//go:build integration
// +build integration

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssuerIntegration(t *testing.T) {
	cli := newConjurTestCLI(t)
	cli.InitAndLoginAsAdmin(t)

	// Create Conjur variables to hold AWS credentials, as required by the
	// server API which no longer accepts inline credential values.
	cli.LoadPolicy(t, `
- !variable aws-credentials/access_key_id
- !variable aws-credentials/secret_access_key
`)
	_, _, err := cli.Run("variable", "set", "-i", "aws-credentials/access_key_id", "-v", "AKIAIOSFODNN7EXAMPLE")
	require.NoError(t, err)
	_, _, err = cli.Run("variable", "set", "-i", "aws-credentials/secret_access_key", "-v", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")
	require.NoError(t, err)

	t.Run("Create issuer", func(t *testing.T) {
		data := `{
      "access_key_id_secret_ref": {"id": "aws-credentials/access_key_id"},
      "secret_access_key_secret_ref": {"id": "aws-credentials/secret_access_key"}
    }`

		stdOut, _, err := cli.Run(
			"issuer", "create",
			"--id", "test-id",
			"--max-ttl", "3000",
			"--type", "aws",
			"--data", data,
		)

		require.NoError(t, err)
		assert.Contains(t, stdOut, `"id": "test-id"`)
		assert.Contains(t, stdOut, `"max_ttl": 3000`)
		assert.Contains(t, stdOut, `"access_key_id_secret_ref"`)
		assert.Contains(t, stdOut, `"secret_access_key_secret_ref"`)
	})

	// Test case ensures server error messages are passed to cli error output
	t.Run("Create issuer with bad id characters", func(t *testing.T) {
		data := `{
      "access_key_id_secret_ref": {"id": "aws-credentials/access_key_id"},
      "secret_access_key_secret_ref": {"id": "aws-credentials/secret_access_key"}
    }`

		_, stdErr, _ := cli.Run(
			"issuer", "create",
			"--id", "t&est-id",
			"--max-ttl", "3000",
			"--type", "aws",
			"--data", data,
		)

		assert.Contains(t, stdErr, "invalid 'id' parameter. Only the following characters are supported: A-Z, a-z, 0-9, +, -, and _.")
	})

	t.Run("List issuers", func(t *testing.T) {

		stdOut, _, _ := cli.Run(
			"issuer", "list",
		)

		assert.Contains(t, stdOut, `"id": "test-id"`)
		assert.Contains(t, stdOut, `"max_ttl": 3000`)
		assert.Contains(t, stdOut, `"access_key_id_secret_ref"`)
		assert.Contains(t, stdOut, `"secret_access_key_secret_ref"`)
	})

	t.Run("Get Issuer", func(t *testing.T) {

		stdOut, _, _ := cli.Run(
			"issuer", "get",
			"--id", "test-id",
		)

		assert.Contains(t, stdOut, `"id": "test-id"`)
		assert.Contains(t, stdOut, `"max_ttl": 3000`)
		assert.Contains(t, stdOut, `"access_key_id_secret_ref"`)
		assert.Contains(t, stdOut, `"secret_access_key_secret_ref"`)
	})

	t.Run("Update Issuer", func(t *testing.T) {

		stdOut, _, _ := cli.Run(
			"issuer", "update",
			"--id", "test-id",
			"--max-ttl", "3001",
		)

		assert.Contains(t, stdOut, `"id": "test-id"`)
		assert.Contains(t, stdOut, `"max_ttl": 3001`)
	})

	t.Run("Attempt to retrieve non-existent issuer", func(t *testing.T) {

		_, stdErr, _ := cli.Run(
			"issuer", "get",
			"--id", "test-id2",
		)

		assert.Contains(t, stdErr, `404 Not Found. Issuer not found.`)
	})

	t.Run("Delete Issuer", func(t *testing.T) {

		stdOut, _, _ := cli.Run(
			"issuer", "delete",
			"--id", "test-id",
		)

		assert.Contains(t, stdOut, `The issuer 'test-id' was deleted`)
	})

	t.Run("Attempt to retrieve deleted issuer", func(t *testing.T) {

		_, stdErr, _ := cli.Run(
			"issuer", "get",
			"--id", "test-id",
		)

		assert.Contains(t, stdErr, `404 Not Found. Issuer not found.`)
	})
}
