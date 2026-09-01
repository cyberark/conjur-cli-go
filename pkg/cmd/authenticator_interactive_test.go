package cmd

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// Cannot run with t.Parallel(): overrides package-level prompting seams.
func TestAuthenticatorCmdInteractivePrompting(t *testing.T) {
	for _, subcommand := range []string{"enable", "disable"} {
		t.Run(subcommand+" subcommand interactive with TTY prompts for missing id", func(t *testing.T) {
			withPromptingSeams(t, true, "authn-iam/prod")

			mockClient := mockAuthenticatorClient{
				t: t,
				enableAuthenticator: func(t *testing.T, authenticatorType string, serviceID string, enabled bool) error {
					return errors.New("stop before output rendering")
				},
			}
			cmd := newAuthenticatorCommand(
				func(cmd *cobra.Command) (authenticatorClient, error) { return mockClient, nil },
			)

			_, stderr, err := executeCommandForTest(t, cmd, "authenticator", subcommand, "--interactive")
			assert.Error(t, err)
			assert.Contains(t, stderr, "stop before output rendering")
		})

		t.Run(subcommand+" subcommand interactive without TTY still fails with required-flag error", func(t *testing.T) {
			withPromptingSeams(t, false)

			mockClient := mockAuthenticatorClient{t: t}
			cmd := newAuthenticatorCommand(
				func(cmd *cobra.Command) (authenticatorClient, error) { return mockClient, nil },
			)

			_, stderr, err := executeCommandForTest(t, cmd, "authenticator", subcommand, "--interactive")
			assert.Error(t, err)
			assert.Contains(t, stderr, "required flag(s)")
		})
	}
}
