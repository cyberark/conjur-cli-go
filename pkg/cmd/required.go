package cmd

import (
	"fmt"

	"github.com/cyberark/conjur-cli-go/pkg/output"
	"github.com/cyberark/conjur-cli-go/pkg/prompts"
	"github.com/spf13/cobra"
)

// promptingAvailable, askForValue, and askForSecretValue are
// test-overridable seams for the prompts package functions of the same
// purpose.
var promptingAvailable = prompts.PromptingAvailable
var askForValue = prompts.AskForValue
var askForSecretValue = prompts.AskForSecretValue

// interactivePromptingEnabled reports whether the command was invoked with
// --interactive (or an equivalent --output interactive alias) AND a real
// TTY is available on both stdin and stdout. Both conditions must hold
// before any required-field prompting is attempted, so that --interactive
// passed in a non-TTY context (CI, a pipe, a background subshell) fails
// fast instead of blocking on a read that will never resolve.
func interactivePromptingEnabled(cmd *cobra.Command) (bool, error) {
	_, isInteractive, err := output.FromContext(cmd)
	if err != nil {
		return false, err
	}
	return isInteractive && promptingAvailable(), nil
}

// requirePositionalArgs ensures exactly len(names) positional args are
// present in args, prompting for any missing ones (in order) when
// interactive prompting is available, or failing with a standard
// Cobra-style arity error otherwise.
func requirePositionalArgs(cmd *cobra.Command, args []string, names ...string) ([]string, error) {
	n := len(names)
	if len(args) == n {
		return args, nil
	}

	arityErr := func() error {
		return fmt.Errorf("accepts %d arg(s), received %d", n, len(args))
	}

	if len(args) > n {
		return nil, arityErr()
	}

	interactive, err := interactivePromptingEnabled(cmd)
	if err != nil {
		return nil, err
	}
	if !interactive {
		return nil, arityErr()
	}

	filled := make([]string, len(args), n)
	copy(filled, args)
	for i := len(filled); i < n; i++ {
		value, err := askForValue(names[i])
		if err != nil {
			return nil, err
		}
		filled = append(filled, value)
	}
	return filled, nil
}

// promptForMissingRequiredFlags fills any of the named required flags the
// caller hasn't explicitly set, when interactive prompting is available.
// Intended to be called from a command's PreRunE, before Cobra's own
// ValidateRequiredFlags runs, so a filled-in flag satisfies that check.
// When interactive prompting isn't available, this is a no-op and Cobra's
// ValidateRequiredFlags produces the normal "required flag(s) ... not set"
// error as it does today.
func promptForMissingRequiredFlags(cmd *cobra.Command, names ...string) error {
	interactive, err := interactivePromptingEnabled(cmd)
	if err != nil {
		return err
	}
	if !interactive {
		return nil
	}

	for _, name := range names {
		if cmd.Flags().Changed(name) {
			continue
		}
		value, err := askForValue(name)
		if err != nil {
			return err
		}
		if err := cmd.Flags().Set(name, value); err != nil {
			return err
		}
	}
	return nil
}
