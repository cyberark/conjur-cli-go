//go:build integration
// +build integration

package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPolicyIntegration(t *testing.T) {
	cli := newConjurTestCLI(t)
	cli.InitAndLoginAsAdmin(t)

	t.Run("load new policy from file", func(t *testing.T) {
		policyText := `
- !policy branch1
`
		tempDir := t.TempDir()
		policyFile := tempDir + "/policy.yml"
		err := os.WriteFile(policyFile, []byte(policyText), 0644)
		assert.NoError(t, err)

		stdOut, stdErr, err := cli.Run("policy", "load", "-b", "root", "-f", policyFile)
		assertPolicyLoadCmd(t, err, stdOut, stdErr)
	})

	t.Run("load new policy from stdin", func(t *testing.T) {
		policyText := `
- !policy branch2
`
		stdOut, stdErr, err := cli.RunWithStdin(
			bytes.NewReader([]byte(policyText)),
			"policy", "load", "-b", "root", "-f", "-",
		)
		assertPolicyLoadCmd(t, err, stdOut, stdErr)
	})

	t.Run("replace policy from stdin", func(t *testing.T) {
		policyText := `
- !policy branch3
`
		stdOut, stdErr, err := cli.RunWithStdin(
			bytes.NewReader([]byte(policyText)),
			"policy", "load", "-b", "branch2", "-f", "-",
		)
		assert.NoError(t, err)
		assert.Contains(t, stdOut, "created_roles")
		assert.Contains(t, stdOut, "version")
		assert.Equal(t, "Loaded policy 'branch2'\n", stdErr)
	})

	t.Run("update policy from stdin", func(t *testing.T) {
		policyText := `
- !policy branch4
`
		stdOut, stdErr, err := cli.RunWithStdin(
			bytes.NewReader([]byte(policyText)),
			"policy", "update", "-b", "branch2", "-f", "-",
		)
		assert.NoError(t, err)
		assert.Contains(t, stdOut, "created_roles")
		assert.Contains(t, stdOut, "version")
		assert.Equal(t, "Loaded policy 'branch2'\n", stdErr)
	})

	t.Run("fetch policy", func(t *testing.T) {
		policyText := `---
- !policy
  id: branch2
  body:
  - !policy
    id: branch3
    body: []
  - !policy
    id: branch4
    body: []

`
		stdOut, _, err := cli.Run(
			"policy", "fetch", "-b", "branch2",
		)
		assert.NoError(t, err)
		assert.NotEmpty(t, stdOut)
		assert.Equal(t, policyText, stdOut)
	})
}
