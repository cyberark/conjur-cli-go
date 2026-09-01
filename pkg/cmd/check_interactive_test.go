package cmd

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// Cannot run with t.Parallel(): overrides package-level prompting seams.
func TestCheckCmdInteractivePrompting(t *testing.T) {
	t.Run("interactive with TTY prompts for missing resource-id and privilege", func(t *testing.T) {
		withPromptingSeams(t, true, "prompted-resource", "prompted-privilege")

		var gotResourceID, gotPrivilege string
		mockClient := mockCheckClient{
			t: t,
			checkPermission: func(t *testing.T, resourceID, privilege string) (bool, error) {
				gotResourceID, gotPrivilege = resourceID, privilege
				return false, errors.New("stop before output rendering")
			},
		}

		cmd := newCheckCmd(func(cmd *cobra.Command) (checkClient, error) { return mockClient, nil })

		_, stderr, err := executeCommandForTest(t, cmd, "check", "--interactive")
		assert.Error(t, err)
		assert.Contains(t, stderr, "stop before output rendering")
		assert.Equal(t, "prompted-resource", gotResourceID)
		assert.Equal(t, "prompted-privilege", gotPrivilege)
	})

	t.Run("interactive without TTY still fails with arity error", func(t *testing.T) {
		withPromptingSeams(t, false)

		cmd := newCheckCmd(func(cmd *cobra.Command) (checkClient, error) { return mockCheckClient{t: t}, nil })

		_, stderr, err := executeCommandForTest(t, cmd, "check", "--interactive")
		assert.Error(t, err)
		assert.Contains(t, stderr, "accepts 2 arg(s), received 0")
	})
}
