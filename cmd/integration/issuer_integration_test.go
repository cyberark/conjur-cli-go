//go:build integration
// +build integration

package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHostIntegration(t *testing.T) {
	cli := newConjurTestCLI(t)
	cli.InitAndLoginAsAdmin(t)

	t.Run("create issuer", func(t *testing.T) {
    data = `{
      "access_key_id": "AKIAIOSFODNN7EXAMPLE",
      "secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKE"
    }`

		stdOut, stdErr, err := cli.Run(
      "issuer", "create",
      "--id", "test-id",
      "--max-ttl", "3000",
      "--type", "aws",
      "--data", data
    )
    
    require.NoError(t, err)
    assert.Contains(t, stdOut, `"id": "test"`)
    assert.Contains(t, stdOut, `"max_ttl": "3000"`)
    assert.Contains(t, stdOut, `"access_key_id": "AKIAIOSFODNN7EXAMPLE"`)
    assert.Contains(t, stdOut, `"secret_access_key": "*****"`)
	})

  // Test case ensure's server error messages are passed to cli error output
	t.Run("create issuer with bad id characters", func(t *testing.T) {
    data = `{
      "access_key_id": "AKIAIOSFODNN7EXAMPLE",
      "secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKE"
    }`

		stdOut, stdErr, err := cli.Run(
      "issuer", "create",
      "--id", "test-id",
      "--max-ttl", "3000",
      "--type", "aws",
      "--data", data
    )
    
    assert.Contains(t, stdErr, "invalid 'id' parameter. Only the following characters are supported: A-Z, a-z, 0-9, +, -, and _.")
	})
}
