package cmd

import (
	"bytes"
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

var assertStopError = errors.New("stop before output rendering")

func executeVariableSetForTest(
	t *testing.T, cmd *cobra.Command, stdin string, args ...string,
) (string, string, error) {
	t.Helper()

	root := newRootCommand()
	root.AddCommand(cmd)
	root.SetArgs(args)

	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)
	mockHelpText(root)
	root.SetOut(stdoutBuf)
	root.SetErr(stderrBuf)
	root.SetIn(bytes.NewReader([]byte(stdin)))

	err := root.Execute()
	return stdoutBuf.String(), stderrBuf.String(), err
}

func TestVariableSetStdinAndTrim(t *testing.T) {
	t.Run("-f - reads value from stdin, raw bytes, no trimming by default", func(t *testing.T) {
		var gotValue string
		mockClient := mockVariableClient{
			t: t,
			set: func(t *testing.T, path, value string) error {
				gotValue = value
				return nil
			},
		}
		cmd := newVariableCmd(
			func(cmd *cobra.Command) (variableGetClient, error) { return mockClient, nil },
			func(cmd *cobra.Command) (variableSetClient, error) { return mockClient, nil },
		)

		_, _, err := executeVariableSetForTest(t, cmd, "piped-secret\n", "variable", "set", "-i", "foo", "-f", "-")
		assert.NoError(t, err)
		assert.Equal(t, "piped-secret\n", gotValue)
	})

	t.Run("-f - with 0 bytes is valid, not an error", func(t *testing.T) {
		var gotValue string
		called := false
		mockClient := mockVariableClient{
			t: t,
			set: func(t *testing.T, path, value string) error {
				called = true
				gotValue = value
				return nil
			},
		}
		cmd := newVariableCmd(
			func(cmd *cobra.Command) (variableGetClient, error) { return mockClient, nil },
			func(cmd *cobra.Command) (variableSetClient, error) { return mockClient, nil },
		)

		_, _, err := executeVariableSetForTest(t, cmd, "", "variable", "set", "-i", "foo", "-f", "-")
		assert.NoError(t, err)
		assert.True(t, called)
		assert.Equal(t, "", gotValue)
	})

	t.Run("--trim-whitespace trims stdin-sourced value", func(t *testing.T) {
		var gotValue string
		mockClient := mockVariableClient{
			t: t,
			set: func(t *testing.T, path, value string) error {
				gotValue = value
				return nil
			},
		}
		cmd := newVariableCmd(
			func(cmd *cobra.Command) (variableGetClient, error) { return mockClient, nil },
			func(cmd *cobra.Command) (variableSetClient, error) { return mockClient, nil },
		)

		_, _, err := executeVariableSetForTest(
			t, cmd, "  piped-secret\n\n", "variable", "set", "-i", "foo", "-f", "-", "--trim-whitespace",
		)
		assert.NoError(t, err)
		assert.Equal(t, "piped-secret", gotValue)
	})

	t.Run("--trim-whitespace trims --value-sourced value", func(t *testing.T) {
		var gotValue string
		mockClient := mockVariableClient{
			t: t,
			set: func(t *testing.T, path, value string) error {
				gotValue = value
				return nil
			},
		}
		cmd := newVariableCmd(
			func(cmd *cobra.Command) (variableGetClient, error) { return mockClient, nil },
			func(cmd *cobra.Command) (variableSetClient, error) { return mockClient, nil },
		)

		_, _, err := executeVariableSetForTest(
			t, cmd, "", "variable", "set", "-i", "foo", "-v", "  spaced value  ", "--trim-whitespace",
		)
		assert.NoError(t, err)
		assert.Equal(t, "spaced value", gotValue)
	})

	t.Run("--trim-whitespace off by default, --value kept raw", func(t *testing.T) {
		var gotValue string
		mockClient := mockVariableClient{
			t: t,
			set: func(t *testing.T, path, value string) error {
				gotValue = value
				return nil
			},
		}
		cmd := newVariableCmd(
			func(cmd *cobra.Command) (variableGetClient, error) { return mockClient, nil },
			func(cmd *cobra.Command) (variableSetClient, error) { return mockClient, nil },
		)

		_, _, err := executeVariableSetForTest(
			t, cmd, "", "variable", "set", "-i", "foo", "-v", "  spaced value  ",
		)
		assert.NoError(t, err)
		assert.Equal(t, "  spaced value  ", gotValue)
	})
}

// Cannot run with t.Parallel(): overrides package-level prompting seams.
func TestVariableSetGetInteractivePrompting(t *testing.T) {
	t.Run("variable get: interactive with TTY prompts for missing --id", func(t *testing.T) {
		withPromptingSeams(t, true, "prompted-id")

		mockClient := mockVariableClient{
			t: t,
			get: func(t *testing.T, path string) ([]byte, error) {
				assert.Equal(t, "prompted-id", path)
				return nil, assertStopError
			},
			getBatch: func(t *testing.T, paths []string) (map[string][]byte, error) {
				assert.Equal(t, []string{"prompted-id"}, paths)
				return nil, assertStopError
			},
		}
		cmd := newVariableCmd(
			func(cmd *cobra.Command) (variableGetClient, error) { return mockClient, nil },
			func(cmd *cobra.Command) (variableSetClient, error) { return mockClient, nil },
		)

		_, stderr, err := executeVariableSetForTest(t, cmd, "", "variable", "get", "--interactive")
		assert.Error(t, err)
		assert.Contains(t, stderr, "stop before output rendering")
	})

	t.Run("variable get: interactive without TTY still fails with required-flag error", func(t *testing.T) {
		withPromptingSeams(t, false)

		mockClient := mockVariableClient{t: t}
		cmd := newVariableCmd(
			func(cmd *cobra.Command) (variableGetClient, error) { return mockClient, nil },
			func(cmd *cobra.Command) (variableSetClient, error) { return mockClient, nil },
		)

		_, stderr, err := executeVariableSetForTest(t, cmd, "", "variable", "get", "--interactive")
		assert.Error(t, err)
		assert.Contains(t, stderr, "required flag(s)")
	})

	t.Run("variable set: interactive with TTY prompts for missing --id", func(t *testing.T) {
		withPromptingSeams(t, true, "prompted-id")

		var gotID string
		mockClient := mockVariableClient{
			t: t,
			set: func(t *testing.T, path, value string) error {
				gotID = path
				return assertStopError
			},
		}
		cmd := newVariableCmd(
			func(cmd *cobra.Command) (variableGetClient, error) { return mockClient, nil },
			func(cmd *cobra.Command) (variableSetClient, error) { return mockClient, nil },
		)

		_, stderr, err := executeVariableSetForTest(t, cmd, "", "variable", "set", "-v", "val", "--interactive")
		assert.Error(t, err)
		assert.Contains(t, stderr, "stop before output rendering")
		assert.Equal(t, "prompted-id", gotID)
	})

	t.Run("variable set: interactive with TTY prompts for missing value (masked)", func(t *testing.T) {
		withPromptingSeams(t, true, "prompted-value")

		var gotValue string
		mockClient := mockVariableClient{
			t: t,
			set: func(t *testing.T, path, value string) error {
				gotValue = value
				return nil
			},
		}
		cmd := newVariableCmd(
			func(cmd *cobra.Command) (variableGetClient, error) { return mockClient, nil },
			func(cmd *cobra.Command) (variableSetClient, error) { return mockClient, nil },
		)

		stdout, _, err := executeVariableSetForTest(t, cmd, "", "variable", "set", "-i", "foo", "--interactive")
		assert.NoError(t, err)
		assert.Equal(t, "prompted-value", gotValue)
		// --interactive requests the interactive TUI, but this test has no
		// real TTY (same as a CI runner or a piped session), so Emit falls
		// back to plain JSON rather than opening /dev/tty.
		assert.JSONEq(t, `{"id":"foo","status":"set"}`, stdout)
	})

	t.Run("variable set: interactive without TTY still fails with exclusivity error, no prompt attempted", func(t *testing.T) {
		withPromptingSeams(t, false)

		mockClient := mockVariableClient{t: t}
		cmd := newVariableCmd(
			func(cmd *cobra.Command) (variableGetClient, error) { return mockClient, nil },
			func(cmd *cobra.Command) (variableSetClient, error) { return mockClient, nil },
		)

		_, stderr, err := executeVariableSetForTest(t, cmd, "", "variable", "set", "-i", "foo", "--interactive")
		assert.Error(t, err)
		assert.Contains(t, stderr, "must specify exactly one of --value or --file")
	})

	t.Run("variable set: --file explicitly given, no value prompt attempted even when interactive", func(t *testing.T) {
		withPromptingSeams(t, true)

		var gotValue string
		mockClient := mockVariableClient{
			t: t,
			set: func(t *testing.T, path, value string) error {
				gotValue = value
				return nil
			},
		}
		cmd := newVariableCmd(
			func(cmd *cobra.Command) (variableGetClient, error) { return mockClient, nil },
			func(cmd *cobra.Command) (variableSetClient, error) { return mockClient, nil },
		)

		stdout, _, err := executeVariableSetForTest(
			t, cmd, "piped-value", "variable", "set", "-i", "foo", "-f", "-", "--interactive",
		)
		assert.NoError(t, err)
		assert.Equal(t, "piped-value", gotValue)
		// See the masked-value test above: no real TTY here, so Emit falls
		// back to plain JSON instead of the interactive TUI.
		assert.JSONEq(t, `{"id":"foo","status":"set"}`, stdout)
	})
}
