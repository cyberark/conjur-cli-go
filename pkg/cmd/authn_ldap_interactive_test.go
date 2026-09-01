package cmd

import (
	"errors"
	"testing"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// These tests exercise authn-ldap's --interactive prompting for required
// flags and positional args. They deliberately make the mocked client
// return an error so RunE returns before reaching output.Plain/Emit --
// --interactive also drives output.Emit's TUI rendering on success, which
// needs a real TTY that isn't available in a unit test. What's captured
// from the mock call (serviceID/groupName/roles) still proves prompting
// filled in the missing values correctly. They cannot run with
// t.Parallel() since they override the package-level
// promptingAvailable/askForValue seams.

func TestLdapGroupsListInteractivePrompting(t *testing.T) {
	t.Run("interactive with TTY prompts for missing --id", func(t *testing.T) {
		withPromptingSeams(t, true, "prompted-id")

		var gotServiceID string
		mockClient := mockLdapGroupClient{
			t: t,
			listGroupMappings: func(t *testing.T, serviceID string) (*conjurapi.LdapMappingListResponse, error) {
				gotServiceID = serviceID
				return nil, errors.New("stop before output rendering")
			},
		}

		cmd := newAuthnLdapCommand(
			func(cmd *cobra.Command) (ldapGroupMappingClient, error) { return mockClient, nil },
			func(cmd *cobra.Command) (ldapUserMappingClient, error) { return mockLdapUserClient{t: t}, nil },
		)

		_, stderr, err := executeCommandForTest(t, cmd, "authn-ldap", "groups", "list", "--interactive")
		assert.Error(t, err)
		assert.Contains(t, stderr, "stop before output rendering")
		assert.Equal(t, "prompted-id", gotServiceID)
	})

	t.Run("interactive without TTY still fails with required-flag error, no prompt attempted", func(t *testing.T) {
		withPromptingSeams(t, false)

		mockClient := mockLdapGroupClient{t: t}

		cmd := newAuthnLdapCommand(
			func(cmd *cobra.Command) (ldapGroupMappingClient, error) { return mockClient, nil },
			func(cmd *cobra.Command) (ldapUserMappingClient, error) { return mockLdapUserClient{t: t}, nil },
		)

		_, stderr, err := executeCommandForTest(t, cmd, "authn-ldap", "groups", "list", "--interactive")
		assert.Error(t, err)
		assert.Contains(t, stderr, "required flag(s)")
	})
}

func TestLdapGroupsCreateInteractivePrompting(t *testing.T) {
	t.Run("interactive with TTY prompts for missing group-name arg and --roles/--id flags", func(t *testing.T) {
		withPromptingSeams(t, true, "prompted-id", "prompted-roles", "prompted-group")

		var gotServiceID, gotGroupName string
		var gotRoles []string
		mockClient := mockLdapGroupClient{
			t: t,
			createGroupMapping: func(t *testing.T, serviceID, groupName string, roles []string) (*conjurapi.LdapGroupMappingResponse, error) {
				gotServiceID, gotGroupName, gotRoles = serviceID, groupName, roles
				return nil, errors.New("stop before output rendering")
			},
		}

		cmd := newAuthnLdapCommand(
			func(cmd *cobra.Command) (ldapGroupMappingClient, error) { return mockClient, nil },
			func(cmd *cobra.Command) (ldapUserMappingClient, error) { return mockLdapUserClient{t: t}, nil },
		)

		_, stderr, err := executeCommandForTest(t, cmd, "authn-ldap", "groups", "create", "--interactive")
		assert.Error(t, err)
		assert.Contains(t, stderr, "stop before output rendering")
		assert.Equal(t, "prompted-id", gotServiceID)
		assert.Equal(t, "prompted-group", gotGroupName)
		assert.Equal(t, []string{"prompted-roles"}, gotRoles)
	})
}
