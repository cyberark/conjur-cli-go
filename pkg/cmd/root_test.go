package cmd

import (
	"bytes"
	"github.com/cyberark/conjur-cli-go/pkg/output"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRootCmd(t *testing.T) {
	t.Run("version flag", func(t *testing.T) {
		rootCmd := newRootCommand()
		stdout, _, err := executeCommandForTest(t, rootCmd, "--version")

		assert.NoError(t, err)
		currentYear := time.Now().Format("2006")
		assert.Equal(t, "Idira™ Secrets Manager CLI version unset-unset\n\nCopyright (c) "+currentYear+" Palo Alto Networks Ltd. All rights reserved.\n<www.cyberark.com>\n", stdout)
	})
}

func TestRootOutputFlagResolution(t *testing.T) {
	t.Run("invalid --output value errors before subcommand runs", func(t *testing.T) {
		cmd := newRootCommand()
		var ran bool
		cmd.AddCommand(&cobra.Command{
			Use: "noop",
			RunE: func(cmd *cobra.Command, args []string) error {
				ran = true
				return nil
			},
		})
		cmd.SetArgs([]string{"--output", "bogus", "noop"})
		cmd.SetOut(new(bytes.Buffer))
		cmd.SetErr(new(bytes.Buffer))

		err := cmd.Execute()
		if err == nil {
			t.Fatalf("expected error for invalid --output value")
		}
		if ran {
			t.Fatalf("subcommand RunE must not run when format resolution fails")
		}
	})

	t.Run("marks --output as explicit only when the flag was passed", func(t *testing.T) {
		var explicit bool
		newNoop := func() *cobra.Command {
			return &cobra.Command{
				Use: "noop",
				RunE: func(cmd *cobra.Command, args []string) error {
					explicit = output.IsExplicit(cmd)
					return nil
				},
			}
		}

		cmd := newRootCommand()
		cmd.AddCommand(newNoop())
		cmd.SetArgs([]string{"noop"})
		cmd.SetOut(new(bytes.Buffer))
		cmd.SetErr(new(bytes.Buffer))
		assert.NoError(t, cmd.Execute())
		assert.False(t, explicit, "explicit should be false when --output is not passed")

		cmd = newRootCommand()
		cmd.AddCommand(newNoop())
		cmd.SetArgs([]string{"--output", "plain", "noop"})
		cmd.SetOut(new(bytes.Buffer))
		cmd.SetErr(new(bytes.Buffer))
		assert.NoError(t, cmd.Execute())
		assert.True(t, explicit, "explicit should be true when --output is passed")
	})
}
