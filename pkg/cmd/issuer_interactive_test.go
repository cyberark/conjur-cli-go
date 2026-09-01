package cmd

import (
	"errors"
	"testing"

	api "github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// Cannot run with t.Parallel(): overrides package-level prompting seams.
func TestCreateIssuerCmdInteractivePrompting(t *testing.T) {
	t.Run("interactive with TTY prompts for missing id/type/max-ttl/data", func(t *testing.T) {
		withPromptingSeams(t, true, "prompted-id", "prompted-type", "3000", `{"k":"v"}`)

		var gotIssuer api.Issuer
		mockClient := &mockIssuerClient{
			createIssuer: func(issuer api.Issuer) (api.Issuer, error) {
				gotIssuer = issuer
				return api.Issuer{}, errors.New("stop before output rendering")
			},
		}

		cmd := newCreateIssuerCmd(func(cmd *cobra.Command) (issuerClient, error) { return mockClient, nil })

		_, stderr, err := executeCommandForTest(t, cmd, "create", "--interactive")
		assert.Error(t, err)
		assert.Contains(t, stderr, "stop before output rendering")
		assert.Equal(t, "prompted-id", gotIssuer.ID)
		assert.Equal(t, "prompted-type", gotIssuer.Type)
		assert.Equal(t, 3000, gotIssuer.MaxTTL)
		assert.Equal(t, "v", gotIssuer.Data["k"])
	})

	t.Run("interactive without TTY still fails with required-flag error", func(t *testing.T) {
		withPromptingSeams(t, false)

		mockClient := &mockIssuerClient{}
		cmd := newCreateIssuerCmd(func(cmd *cobra.Command) (issuerClient, error) { return mockClient, nil })

		_, stderr, err := executeCommandForTest(t, cmd, "create", "--interactive")
		assert.Error(t, err)
		assert.Contains(t, stderr, "required flag(s)")
	})
}
