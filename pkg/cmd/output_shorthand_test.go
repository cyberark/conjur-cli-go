package cmd

import (
	"bytes"
	"testing"

	"github.com/cyberark/conjur-cli-go/pkg/output"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// execRootForShorthandTest runs adjustOutputShorthand exactly as Execute()
// does, then executes cmd on a fresh root command tree containing a noop
// leaf and a stand-in for list's -o/--offset collision, and reports
// whether execution panicked.
func execRootForShorthandTest(t *testing.T, args []string) (panicked bool, resolvedFormat output.Format) {
	t.Helper()

	root := newRootCommand()

	noop := &cobra.Command{
		Use: "noop",
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedFormat, _, _ = output.FromContext(cmd)
			return nil
		},
	}
	root.AddCommand(noop)

	listStandIn := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedFormat, _, _ = output.FromContext(cmd)
			return nil
		},
	}
	listStandIn.Flags().IntP("offset", "o", 0, "stand-in for list's real --offset")
	listStandIn.Flags().IntP("limit", "l", 0, "stand-in for list's real --limit")
	root.AddCommand(listStandIn)

	root.SetOut(new(bytes.Buffer))
	root.SetErr(new(bytes.Buffer))
	root.SetArgs(args)

	defer func() {
		if r := recover(); r != nil {
			panicked = true
		}
	}()
	adjustOutputShorthand(root, args)
	_ = root.Execute()
	return panicked, resolvedFormat
}

func TestAdjustOutputShorthandNeverPanics(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"list with its own -o offset", []string{"list", "-o", "5"}},
		{"-o before the list token, meant as offset (positionally passes through to list)", []string{"-o", "5", "list"}},
		{"--offset long form before the list token", []string{"--offset", "5", "list"}},
		{"-o meant as output on a non-colliding command", []string{"noop", "-o", "json"}},
		{"--output long form", []string{"noop", "--output", "json"}},
		{"--output= inline form", []string{"noop", "--output=json"}},
		{"global flag before the subcommand", []string{"--debug", "list", "-o", "5"}},
		{"value-taking global flag before the subcommand", []string{"--timeout", "5s", "list", "--limit", "1"}},
		{"no args at all", []string{}},
		{"unknown top-level flag", []string{"--nope", "list"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			panicked, _ := execRootForShorthandTest(t, tc.args)
			assert.False(t, panicked, "must never panic regardless of argv shape")
		})
	}
}

// TestAdjustOutputShorthandAgainstRealCommandTree exercises the actual
// production rootCmd (all real subcommands, including the real list with
// its real --offset), using Find alone (no RunE/network calls) to detect
// the exact panic class this whole mechanism exists to avoid — this is
// the regression test for the -o/--offset collision this file works
// around, against the real tree rather than a stand-in.
func TestAdjustOutputShorthandAgainstRealCommandTree(t *testing.T) {
	cases := [][]string{
		{"list", "-o", "5"},
		{"-o", "5", "list"},
		{"--offset", "5", "list"},
		{"whoami", "-o", "json"},
		{"variable", "get", "-o", "json"},
		{"authn-ldap", "--id", "x", "groups", "list", "-o", "json"},
		{"policy", "fetch", "-b", "root", "-o", "json"},
		{"resource", "exists", "dev:variable:x", "-o", "json"},
	}

	for _, args := range cases {
		t.Run(joinArgs(args), func(t *testing.T) {
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("panicked for args %v: %v", args, r)
					}
				}()
				adjustOutputShorthand(rootCmd, args)
				_, _, _ = rootCmd.Find(args)
			}()
		})
	}
}

func joinArgs(args []string) string {
	out := ""
	for i, a := range args {
		if i > 0 {
			out += "_"
		}
		out += a
	}
	return out
}

func TestAdjustOutputShorthandResolvesCorrectly(t *testing.T) {
	t.Run("-o means output on a command with no local -o", func(t *testing.T) {
		_, format := execRootForShorthandTest(t, []string{"noop", "-o", "json"})
		assert.Equal(t, output.FormatJSON, format.Type)
	})

	t.Run("--output still works on list itself", func(t *testing.T) {
		_, format := execRootForShorthandTest(t, []string{"list", "--output", "json"})
		assert.Equal(t, output.FormatJSON, format.Type)
	})
}
