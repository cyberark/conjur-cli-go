//go:build integration
// +build integration

package main

import (
	"bytes"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolicyValidationIntegration(t *testing.T) {
	cli := newConjurTestCLI(t)
	cli.InitAndLoginAsAdmin(t)

	t.Run("validate new policy from file", func(t *testing.T) {
		policyText := `
- !policy branch1
`
		tempDir := t.TempDir()
		policyFile := tempDir + "/policy.yml"
		err := os.WriteFile(policyFile, []byte(policyText), 0644)
		require.NoError(t, err)

		stdOut, stdErr, err := cli.Run("policy", "load", "-b", "root", "--dry-run", "-f", policyFile)
		assertPolicyValidateSuccessCmd(t, err, stdOut, stdErr)
	})

	t.Run("validate new policy from stdin", func(t *testing.T) {
		policyText := `
- !policy branch2
`
		stdOut, stdErr, err := cli.RunWithStdin(
			bytes.NewReader([]byte(policyText)),
			"policy", "load", "-b", "root", "--dry-run", "-f", "-",
		)
		assertPolicyValidateSuccessCmd(t, err, stdOut, stdErr)
	})

	// The policyN statements are valid but not executed during dry-run,
	// thus the branches are not created and load operations on them are invalid.
	t.Run("replace policy with validation from stdin", func(t *testing.T) {
		policyText := `
- !policy branch3
`
		stdOut, stdErr, _ := cli.RunWithStdin(
			bytes.NewReader([]byte(policyText)),
			"policy", "load", "-b", "branch2", "--dry-run", "-f", "-",
		)
		require.Contains(t, stdErr, "404 Not Found.")
		assert.Contains(t, stdOut, "")
	})

	// Same as for replace.
	t.Run("update policy with validation from stdin", func(t *testing.T) {
		policyText := `
- !policy branch4
`
		stdOut, stdErr, _ := cli.RunWithStdin(
			bytes.NewReader([]byte(policyText)),
			"policy", "update", "-b", "branch2", "--dry-run", "-f", "-",
		)
		require.Contains(t, stdErr, "404 Not Found.")
		assert.Contains(t, stdOut, "")
	})

	t.Run("validate a policy containing a YAML error", func(t *testing.T) {
		policyText := `
- !!str, xxx
`
		tempDir := t.TempDir()
		policyFile := tempDir + "/policy.yml"
		err := os.WriteFile(policyFile, []byte(policyText), 0644)
		require.NoError(t, err)

		stdOut, stdErr, err := cli.Run("policy", "load", "-b", "root", "--dry-run", "-f", policyFile)
		assertPolicyValidateInvalidCmd(t, err, stdOut, stdErr)
		assert.Contains(t, stdOut, "Invalid YAML")
	})

	t.Run("validate a policy containing a Conjur policy error", func(t *testing.T) {
		policyText := `
- user Alice
`
		tempDir := t.TempDir()
		policyFile := tempDir + "/policy.yml"
		err := os.WriteFile(policyFile, []byte(policyText), 0644)
		require.NoError(t, err)

		stdOut, stdErr, err := cli.Run("policy", "load", "-b", "root", "--dry-run", "-f", policyFile)
		assertPolicyValidateInvalidCmd(t, err, stdOut, stdErr)
		assert.Contains(t, stdOut, "undefined method")
	})

	// This test is brittle given that
	// 1. the expectation policyResponse contains lots of whitespace, and
	// 2. it contains error 'advice' in the message field that could conceivably be update in the future.
	t.Run("validate the entire validation response structure of a policy error", func(t *testing.T) {
		// The policyResponse is sensitive to the newlines in the policyText so
		// use caution if you refactor the declaration.
		policyText :=
			`- !user alice
- !user bob
- !policy
  id: test
  body
  - !user me
`
		policyResponse := `(?s)"status": "Invalid YAML",.*"errors":.*"line": 5,.*"column": 3,.*"message": "could not find expected ':' while scanning a simple key.*This error can occur when you have a missing ':' or missing space after ':'`

		tempDir := t.TempDir()
		policyFile := tempDir + "/policy.yml"
		err := os.WriteFile(policyFile, []byte(policyText), 0644)
		require.NoError(t, err)

		stdOut, stdErr, err := cli.Run("policy", "load", "-b", "root", "--dry-run", "-f", policyFile)
		assertPolicyValidateInvalidCmd(t, err, stdOut, stdErr)
		assert.Regexp(t, regexp.MustCompile(policyResponse), stdOut)
	})

	t.Run("dry run load of a new policy returns content in the Created section", func(t *testing.T) {
		// Load empty policy to clear
		cli.LoadPolicy(t, emptyPolicy)

		// Load new policy to Create
		stdOut, stdErr, err := cli.DryRunPolicy(t, "load", "root", testPolicy)
		require.NoError(t, err)
		assert.Contains(t, stdErr, "Dry run policy 'root'")
		assert.Regexp(t, regexp.MustCompile(`(?s)created`), stdOut)
	})

	t.Run("dry run update of an existing policy returns content in the Updated sections", func(t *testing.T) {
		// Load test policy as baseline
		cli.LoadPolicy(t, testPolicy)

		// Update policy
		stdOut, stdErr, err := cli.DryRunPolicy(t, "update", "root",
			`
- !permit
  role: !user alice
  resource: !variable woof
  privileges: [ write ]
`)
		require.NoError(t, err)
		assert.Contains(t, stdErr, "Dry run policy 'root'")
		assert.Regexp(t, regexp.MustCompile(`(?s)created.*updated.*before.*items.*permissions.*variable:meow.*after.*items.*permissions.*variable:meow.*variable:woof`), stdOut)
	})

	t.Run("dry run replace of an existing policy returns content in the Deleted section", func(t *testing.T) {
		// Load test policy as baseline
		cli.LoadPolicy(t, testPolicy)

		// Replace policy
		stdOut, stdErr, err := cli.DryRunPolicy(t, "replace", "root", emptyPolicy)
		require.NoError(t, err)
		assert.Contains(t, stdErr, "Dry run policy 'root'")
		assert.Regexp(t, regexp.MustCompile(`(?s)status.*deleted`), stdOut)
	})

	t.Run("dry run load against an existing policy returns content in all sections", func(t *testing.T) {
		// Load test policy as baseline
		cli.LoadPolicy(t, testPolicy)

		// Load a policy causing all types of changes
		stdOut, stdErr, err := cli.DryRunPolicy(t, "replace", "root",
			`
- !user charles
- !delete
  record: !variable meow
- !variable blub
            `)
		require.NoError(t, err)
		assert.Contains(t, stdErr, "Dry run policy 'root'")
		assert.Regexp(t, regexp.MustCompile(`(?s)status.*created.*updated.*deleted.*errors`), stdOut)
	})
}
