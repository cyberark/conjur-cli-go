package cmd

import (
	"bytes"
	"testing"

	"github.com/cyberark/conjur-cli-go/pkg/output"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func execRootWithNoop(t *testing.T, args []string, capture func(cmd *cobra.Command)) error {
	t.Helper()
	cmd := newRootCommand()
	cmd.AddCommand(&cobra.Command{
		Use: "noop",
		RunE: func(cmd *cobra.Command, args []string) error {
			capture(cmd)
			return nil
		},
	})
	cmd.SetArgs(append([]string{"noop"}, args...))
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	return cmd.Execute()
}

func TestOutputAliasFlags(t *testing.T) {
	t.Run("--json resolves to output json", func(t *testing.T) {
		var got output.Format
		err := execRootWithNoop(t, []string{"--json"}, func(cmd *cobra.Command) {
			got, _, _ = output.FromContext(cmd)
		})
		assert.NoError(t, err)
		assert.Equal(t, output.FormatJSON, got.Type)
	})

	t.Run("--plain is explicit, unlike omitting --output", func(t *testing.T) {
		var explicit bool
		err := execRootWithNoop(t, []string{"--plain"}, func(cmd *cobra.Command) {
			explicit = output.IsExplicit(cmd)
		})
		assert.NoError(t, err)
		assert.True(t, explicit)
	})

	t.Run("--columns builds a custom-columns spec from bare field names", func(t *testing.T) {
		var got output.Format
		err := execRootWithNoop(t, []string{"--columns=id,kind"}, func(cmd *cobra.Command) {
			got, _, _ = output.FromContext(cmd)
		})
		assert.NoError(t, err)
		assert.Equal(t, output.FormatCustomColumns, got.Type)
		assert.Equal(t, []output.ColumnSpec{{Header: "id", Path: ".id"}, {Header: "kind", Path: ".kind"}}, got.Columns)
	})

	t.Run("--template builds a go-template spec", func(t *testing.T) {
		var got output.Format
		err := execRootWithNoop(t, []string{"--template={{.name}}"}, func(cmd *cobra.Command) {
			got, _, _ = output.FromContext(cmd)
		})
		assert.NoError(t, err)
		assert.Equal(t, output.FormatGoTemplate, got.Type)
	})

	t.Run("combining two aliases errors", func(t *testing.T) {
		err := execRootWithNoop(t, []string{"--json", "--yaml"}, func(cmd *cobra.Command) {
			t.Fatalf("RunE must not run when aliases conflict")
		})
		assert.Error(t, err)
	})

	t.Run("--output and a conflicting alias errors", func(t *testing.T) {
		err := execRootWithNoop(t, []string{"--output", "yaml", "--json"}, func(cmd *cobra.Command) {
			t.Fatalf("RunE must not run when --output conflicts with an alias")
		})
		assert.Error(t, err)
	})

	t.Run("--output and a matching alias is not a conflict", func(t *testing.T) {
		var got output.Format
		err := execRootWithNoop(t, []string{"--output", "json", "--json"}, func(cmd *cobra.Command) {
			got, _, _ = output.FromContext(cmd)
		})
		assert.NoError(t, err)
		assert.Equal(t, output.FormatJSON, got.Type)
	})

	t.Run("no flags at all resolves to plain, not explicit", func(t *testing.T) {
		var got output.Format
		var explicit bool
		err := execRootWithNoop(t, nil, func(cmd *cobra.Command) {
			got, _, _ = output.FromContext(cmd)
			explicit = output.IsExplicit(cmd)
		})
		assert.NoError(t, err)
		assert.Equal(t, output.FormatPlain, got.Type)
		assert.False(t, explicit)
	})
}

func TestOutputAliasDoesNotShadowASubcommandsOwnFlagOfTheSameName(t *testing.T) {
	// A subcommand defining its own local flag with the same name as one
	// of the output aliases (e.g. --interactive) must keep its own,
	// unrelated meaning rather than being reinterpreted as an
	// output-format request.
	localCmd := &cobra.Command{Use: "local"}
	localCmd.Flags().Bool("interactive", true, "unrelated local meaning")

	root := newRootCommand()
	root.AddCommand(localCmd)

	assert.False(t, isPersistentFlagInEffect(localCmd, "interactive"))
}
