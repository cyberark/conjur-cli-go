package cmd

import (
	"testing"

	"github.com/cyberark/conjur-cli-go/pkg/output"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func newTestCmdWithOutputContext(isInteractive bool) *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	format, _, _ := output.Resolve("")
	output.SetContext(cmd, format, isInteractive)
	return cmd
}

func withPromptingSeams(t *testing.T, available bool, values ...string) {
	t.Helper()
	origAvailable := promptingAvailable
	origAsk := askForValue
	origAskSecret := askForSecretValue
	promptingAvailable = func() bool { return available }
	i := 0
	next := func(name string) (string, error) {
		v := values[i]
		i++
		return v, nil
	}
	askForValue = next
	askForSecretValue = next
	t.Cleanup(func() {
		promptingAvailable = origAvailable
		askForValue = origAsk
		askForSecretValue = origAskSecret
	})
}

func TestRequirePositionalArgs(t *testing.T) {
	t.Run("all args present, no prompting attempted", func(t *testing.T) {
		cmd := newTestCmdWithOutputContext(false)
		args, err := requirePositionalArgs(cmd, []string{"a", "b"}, "first", "second")
		assert.NoError(t, err)
		assert.Equal(t, []string{"a", "b"}, args)
	})

	t.Run("too many args, non-interactive, fails with arity error", func(t *testing.T) {
		cmd := newTestCmdWithOutputContext(false)
		_, err := requirePositionalArgs(cmd, []string{"a", "b", "c"}, "first", "second")
		assert.ErrorContains(t, err, "accepts 2 arg(s), received 3")
	})

	t.Run("missing arg, non-interactive, fails with arity error", func(t *testing.T) {
		cmd := newTestCmdWithOutputContext(false)
		_, err := requirePositionalArgs(cmd, []string{"a"}, "first", "second")
		assert.ErrorContains(t, err, "accepts 2 arg(s), received 1")
	})

	t.Run("missing arg, interactive but no TTY, fails with arity error", func(t *testing.T) {
		withPromptingSeams(t, false)
		cmd := newTestCmdWithOutputContext(true)
		_, err := requirePositionalArgs(cmd, []string{"a"}, "first", "second")
		assert.ErrorContains(t, err, "accepts 2 arg(s), received 1")
	})

	t.Run("missing arg, interactive with TTY, prompts and fills in order", func(t *testing.T) {
		withPromptingSeams(t, true, "prompted-second")
		cmd := newTestCmdWithOutputContext(true)
		args, err := requirePositionalArgs(cmd, []string{"a"}, "first", "second")
		assert.NoError(t, err)
		assert.Equal(t, []string{"a", "prompted-second"}, args)
	})

	t.Run("all args missing, interactive with TTY, prompts for each in order", func(t *testing.T) {
		withPromptingSeams(t, true, "prompted-first", "prompted-second")
		cmd := newTestCmdWithOutputContext(true)
		args, err := requirePositionalArgs(cmd, []string{}, "first", "second")
		assert.NoError(t, err)
		assert.Equal(t, []string{"prompted-first", "prompted-second"}, args)
	})
}

func TestPromptForMissingRequiredFlags(t *testing.T) {
	t.Run("non-interactive, no-op, flag left unset", func(t *testing.T) {
		cmd := newTestCmdWithOutputContext(false)
		cmd.Flags().String("id", "", "")
		err := promptForMissingRequiredFlags(cmd, "id")
		assert.NoError(t, err)
		assert.False(t, cmd.Flags().Changed("id"))
	})

	t.Run("interactive but no TTY, no-op, flag left unset", func(t *testing.T) {
		withPromptingSeams(t, false)
		cmd := newTestCmdWithOutputContext(true)
		cmd.Flags().String("id", "", "")
		err := promptForMissingRequiredFlags(cmd, "id")
		assert.NoError(t, err)
		assert.False(t, cmd.Flags().Changed("id"))
	})

	t.Run("interactive with TTY, missing flag gets prompted and set", func(t *testing.T) {
		withPromptingSeams(t, true, "prompted-id")
		cmd := newTestCmdWithOutputContext(true)
		cmd.Flags().String("id", "", "")
		err := promptForMissingRequiredFlags(cmd, "id")
		assert.NoError(t, err)
		assert.True(t, cmd.Flags().Changed("id"))
		v, _ := cmd.Flags().GetString("id")
		assert.Equal(t, "prompted-id", v)
	})

	t.Run("interactive with TTY, explicitly-set empty flag is not re-prompted", func(t *testing.T) {
		withPromptingSeams(t, true)
		cmd := newTestCmdWithOutputContext(true)
		cmd.Flags().String("id", "", "")
		err := cmd.Flags().Set("id", "")
		assert.NoError(t, err)
		err = promptForMissingRequiredFlags(cmd, "id")
		assert.NoError(t, err)
		v, _ := cmd.Flags().GetString("id")
		assert.Equal(t, "", v)
	})
}
