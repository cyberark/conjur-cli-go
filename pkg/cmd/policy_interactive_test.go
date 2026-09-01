package cmd

import (
	"errors"
	"io"
	"testing"

	api "github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// Cannot run with t.Parallel(): overrides package-level prompting seams.
func TestPolicyLoadCmdInteractivePrompting(t *testing.T) {
	for _, subcommand := range []string{"load", "update", "replace"} {
		t.Run(subcommand+" subcommand interactive with TTY prompts for missing branch/file", func(t *testing.T) {
			withPromptingSeams(t, true, "prompted-branch", "-")

			mockClient := mockPolicyClient{
				t: t,
				loadPolicy: func(t *testing.T, mode api.PolicyMode, policyBranch string, policySrc io.Reader) (*api.PolicyResponse, error) {
					return nil, errors.New("stop before output rendering")
				},
			}
			cmd := newPolicyCommand(
				func(cmd *cobra.Command) (policyClient, error) { return mockClient, nil },
			)

			_, stderr, err := executeCommandForTest(t, cmd, "policy", subcommand, "--interactive")
			assert.Error(t, err)
			assert.Contains(t, stderr, "stop before output rendering")
		})

		t.Run(subcommand+" subcommand interactive without TTY still fails with required-flag error", func(t *testing.T) {
			withPromptingSeams(t, false)

			mockClient := mockPolicyClient{t: t}
			cmd := newPolicyCommand(
				func(cmd *cobra.Command) (policyClient, error) { return mockClient, nil },
			)

			_, stderr, err := executeCommandForTest(t, cmd, "policy", subcommand, "--interactive")
			assert.Error(t, err)
			assert.Contains(t, stderr, "required flag(s)")
		})
	}
}

func TestPolicyFetchCmdInteractivePrompting(t *testing.T) {
	t.Run("fetch subcommand interactive with TTY prompts for missing branch", func(t *testing.T) {
		withPromptingSeams(t, true, "prompted-branch")

		mockClient := mockPolicyClient{
			t: t,
			fetchPolicy: func(t *testing.T, policyBranch string, returnJSON bool, policyTreeDepth uint, sizeLimit uint) ([]byte, error) {
				return nil, errors.New("stop before output rendering")
			},
		}
		cmd := newPolicyCommand(
			func(cmd *cobra.Command) (policyClient, error) { return mockClient, nil },
		)

		_, stderr, err := executeCommandForTest(t, cmd, "policy", "fetch", "--interactive")
		assert.Error(t, err)
		assert.Contains(t, stderr, "stop before output rendering")
	})

	t.Run("fetch subcommand interactive without TTY still fails with required-flag error", func(t *testing.T) {
		withPromptingSeams(t, false)

		mockClient := mockPolicyClient{t: t}
		cmd := newPolicyCommand(
			func(cmd *cobra.Command) (policyClient, error) { return mockClient, nil },
		)

		_, stderr, err := executeCommandForTest(t, cmd, "policy", "fetch", "--interactive")
		assert.Error(t, err)
		assert.Contains(t, stderr, "required flag(s)")
	})
}
