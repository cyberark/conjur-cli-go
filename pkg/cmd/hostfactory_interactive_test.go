package cmd

import (
	"errors"
	"testing"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// Cannot run with t.Parallel(): overrides package-level prompting seams.
func TestHostsCreateCmdInteractivePrompting(t *testing.T) {
	t.Run("interactive with TTY prompts for missing --token and --id", func(t *testing.T) {
		withPromptingSeams(t, true, "prompted-token", "prompted-id")

		var gotID, gotToken string
		mockClient := mockHFClient{
			t: t,
			host: func(t *testing.T, id, token string) (conjurapi.HostFactoryHostResponse, error) {
				gotID, gotToken = id, token
				return conjurapi.HostFactoryHostResponse{}, errors.New("stop before output rendering")
			},
		}

		cmd := newHostsCreateCmd(func(cmd *cobra.Command) (createHostClient, error) { return mockClient, nil })

		_, stderr, err := executeCommandForTest(t, cmd, "create", "--interactive")
		assert.Error(t, err)
		assert.Contains(t, stderr, "stop before output rendering")
		assert.Equal(t, "prompted-id", gotID)
		assert.Equal(t, "prompted-token", gotToken)
	})

	t.Run("interactive without TTY still fails with required-flag error", func(t *testing.T) {
		withPromptingSeams(t, false)

		mockClient := mockHFClient{t: t}
		cmd := newHostsCreateCmd(func(cmd *cobra.Command) (createHostClient, error) { return mockClient, nil })

		_, stderr, err := executeCommandForTest(t, cmd, "create", "--interactive")
		assert.Error(t, err)
		assert.Contains(t, stderr, "required flag(s)")
	})
}
