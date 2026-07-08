package cmd

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// --- Group mock ---

type mockLdapGroupClient struct {
	t                  *testing.T
	createGroupMapping func(*testing.T, string, string, []string) (*conjurapi.LdapGroupMappingResponse, error)
	showGroupMapping   func(*testing.T, string, string) (*conjurapi.LdapGroupMappingResponse, error)
	listGroupMappings  func(*testing.T, string) (*conjurapi.LdapMappingListResponse, error)
	deleteGroupMapping func(*testing.T, string, string) error
}

func (m mockLdapGroupClient) LdapCreateGroupMapping(serviceID, groupName string, roles []string) (*conjurapi.LdapGroupMappingResponse, error) {
	return m.createGroupMapping(m.t, serviceID, groupName, roles)
}

func (m mockLdapGroupClient) LdapShowGroupMapping(serviceID, groupName string) (*conjurapi.LdapGroupMappingResponse, error) {
	return m.showGroupMapping(m.t, serviceID, groupName)
}

func (m mockLdapGroupClient) LdapListGroupMappings(serviceID string) (*conjurapi.LdapMappingListResponse, error) {
	return m.listGroupMappings(m.t, serviceID)
}

func (m mockLdapGroupClient) LdapDeleteGroupMapping(serviceID, groupName string) error {
	return m.deleteGroupMapping(m.t, serviceID, groupName)
}

// --- User mock ---

type mockLdapUserClient struct {
	t                  *testing.T
	createUserMapping  func(*testing.T, string, string, []string) (*conjurapi.LdapUserMappingResponse, error)
	showUserMapping    func(*testing.T, string, string) (*conjurapi.LdapUserMappingResponse, error)
	listUserMappings   func(*testing.T, string) (*conjurapi.LdapMappingListResponse, error)
	deleteUserMapping  func(*testing.T, string, string) error
}

func (m mockLdapUserClient) LdapCreateUserMapping(serviceID, username string, roles []string) (*conjurapi.LdapUserMappingResponse, error) {
	return m.createUserMapping(m.t, serviceID, username, roles)
}

func (m mockLdapUserClient) LdapShowUserMapping(serviceID, username string) (*conjurapi.LdapUserMappingResponse, error) {
	return m.showUserMapping(m.t, serviceID, username)
}

func (m mockLdapUserClient) LdapListUserMappings(serviceID string) (*conjurapi.LdapMappingListResponse, error) {
	return m.listUserMappings(m.t, serviceID)
}

func (m mockLdapUserClient) LdapDeleteUserMapping(serviceID, username string) error {
	return m.deleteUserMapping(m.t, serviceID, username)
}

// --- Group test cases ---

type ldapGroupCmdTestCase struct {
	name               string
	args               []string
	createGroupMapping func(*testing.T, string, string, []string) (*conjurapi.LdapGroupMappingResponse, error)
	showGroupMapping   func(*testing.T, string, string) (*conjurapi.LdapGroupMappingResponse, error)
	listGroupMappings  func(*testing.T, string) (*conjurapi.LdapMappingListResponse, error)
	deleteGroupMapping func(*testing.T, string, string) error
	clientFactoryError error
	assert             func(*testing.T, string, string, error)
}

var ldapGroupCmdTestCases = []ldapGroupCmdTestCase{
	{
		name: "groups create success table",
		args: []string{"authn-ldap", "--id", "ad-prod", "groups", "create", "devs", "--roles", "dev-role"},
		createGroupMapping: func(t *testing.T, serviceID, groupName string, roles []string) (*conjurapi.LdapGroupMappingResponse, error) {
			assert.Equal(t, "ad-prod", serviceID)
			assert.Equal(t, "devs", groupName)
			assert.Equal(t, []string{"dev-role"}, roles)
			return &conjurapi.LdapGroupMappingResponse{LdapGroup: "devs", Roles: []string{"dev-role"}}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.NoError(t, err)
			assert.Empty(t, stderr)
			assert.Contains(t, stdout, "devs")
			assert.Contains(t, stdout, "dev-role")
		},
	},
	{
		name: "groups create success json",
		args: []string{"authn-ldap", "--id", "ad-prod", "--output", "json", "groups", "create", "devs", "--roles", "dev-role"},
		createGroupMapping: func(t *testing.T, serviceID, groupName string, roles []string) (*conjurapi.LdapGroupMappingResponse, error) {
			return &conjurapi.LdapGroupMappingResponse{LdapGroup: "devs", Roles: []string{"dev-role"}}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.NoError(t, err)
			var result conjurapi.LdapGroupMappingResponse
			assert.NoError(t, json.Unmarshal([]byte(stdout), &result))
			assert.Equal(t, "devs", result.LdapGroup)
			assert.Equal(t, []string{"dev-role"}, result.Roles)
		},
	},
	{
		name: "groups create API error",
		args: []string{"authn-ldap", "--id", "ad-prod", "groups", "create", "devs", "--roles", "dev-role"},
		createGroupMapping: func(t *testing.T, serviceID, groupName string, roles []string) (*conjurapi.LdapGroupMappingResponse, error) {
			return nil, errors.New("404 Not Found")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to create group mapping 'devs'")
			assert.Contains(t, err.Error(), "404 Not Found")
		},
	},
	{
		name:               "groups create client factory error",
		args:               []string{"authn-ldap", "--id", "ad-prod", "groups", "create", "devs", "--roles", "dev-role"},
		clientFactoryError: errors.New("client factory error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, stderr, "Error: client factory error")
		},
	},
	{
		name: "groups show success",
		args: []string{"authn-ldap", "--id", "ad-prod", "groups", "show", "admins"},
		showGroupMapping: func(t *testing.T, serviceID, groupName string) (*conjurapi.LdapGroupMappingResponse, error) {
			assert.Equal(t, "ad-prod", serviceID)
			assert.Equal(t, "admins", groupName)
			return &conjurapi.LdapGroupMappingResponse{LdapGroup: "admins", Roles: []string{"admin-role"}}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.NoError(t, err)
			assert.Contains(t, stdout, "admins")
			assert.Contains(t, stdout, "admin-role")
		},
	},
	{
		name: "groups show API error",
		args: []string{"authn-ldap", "--id", "ad-prod", "groups", "show", "admins"},
		showGroupMapping: func(t *testing.T, serviceID, groupName string) (*conjurapi.LdapGroupMappingResponse, error) {
			return nil, errors.New("404 Not Found")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to show group mapping 'admins'")
		},
	},
	{
		name: "groups list success table",
		args: []string{"authn-ldap", "--id", "ad-prod", "groups", "list"},
		listGroupMappings: func(t *testing.T, serviceID string) (*conjurapi.LdapMappingListResponse, error) {
			assert.Equal(t, "ad-prod", serviceID)
			return &conjurapi.LdapMappingListResponse{Keys: []string{"admins", "devs"}}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.NoError(t, err)
			assert.Contains(t, stdout, "admins")
			assert.Contains(t, stdout, "devs")
		},
	},
	{
		name: "groups list success json",
		args: []string{"authn-ldap", "--id", "ad-prod", "--output", "json", "groups", "list"},
		listGroupMappings: func(t *testing.T, serviceID string) (*conjurapi.LdapMappingListResponse, error) {
			return &conjurapi.LdapMappingListResponse{Keys: []string{"admins", "devs"}}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.NoError(t, err)
			var result conjurapi.LdapMappingListResponse
			assert.NoError(t, json.Unmarshal([]byte(stdout), &result))
			assert.Equal(t, []string{"admins", "devs"}, result.Keys)
		},
	},
	{
		name: "groups list API error",
		args: []string{"authn-ldap", "--id", "ad-prod", "groups", "list"},
		listGroupMappings: func(t *testing.T, serviceID string) (*conjurapi.LdapMappingListResponse, error) {
			return nil, errors.New("403 Forbidden")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to list group mappings")
		},
	},
	{
		name: "groups delete success",
		args: []string{"authn-ldap", "--id", "ad-prod", "groups", "delete", "admins"},
		deleteGroupMapping: func(t *testing.T, serviceID, groupName string) error {
			assert.Equal(t, "ad-prod", serviceID)
			assert.Equal(t, "admins", groupName)
			return nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.NoError(t, err)
			assert.Contains(t, stdout, "Success! LDAP group mapping 'admins' deleted")
		},
	},
	{
		name: "groups delete API error",
		args: []string{"authn-ldap", "--id", "ad-prod", "groups", "delete", "admins"},
		deleteGroupMapping: func(t *testing.T, serviceID, groupName string) error {
			return errors.New("404 Not Found")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to delete group mapping 'admins'")
		},
	},
	{
		name: "groups list invalid output format",
		args: []string{"authn-ldap", "--id", "ad-prod", "--output", "xml", "groups", "list"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid output format")
		},
	},
	{
		name: "groups list missing id flag",
		args: []string{"authn-ldap", "groups", "list"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, stderr, "required flag(s)")
		},
	},
}

func TestLdapGroupsCmd(t *testing.T) {
	t.Parallel()
	for _, tc := range ldapGroupCmdTestCases {
		t.Run(tc.name, func(t *testing.T) {
			mockGroup := mockLdapGroupClient{
				t:                  t,
				createGroupMapping: tc.createGroupMapping,
				showGroupMapping:   tc.showGroupMapping,
				listGroupMappings:  tc.listGroupMappings,
				deleteGroupMapping: tc.deleteGroupMapping,
			}
			mockUser := mockLdapUserClient{t: t}

			cmd := newAuthnLdapCommand(
				func(_ *cobra.Command) (ldapGroupMappingClient, error) {
					return mockGroup, tc.clientFactoryError
				},
				func(_ *cobra.Command) (ldapUserMappingClient, error) {
					return mockUser, nil
				},
			)

			stdout, stderr, err := executeCommandForTest(t, cmd, tc.args...)
			tc.assert(t, stdout, stderr, err)
		})
	}
}

// --- User test cases ---

type ldapUserCmdTestCase struct {
	name               string
	args               []string
	createUserMapping  func(*testing.T, string, string, []string) (*conjurapi.LdapUserMappingResponse, error)
	showUserMapping    func(*testing.T, string, string) (*conjurapi.LdapUserMappingResponse, error)
	listUserMappings   func(*testing.T, string) (*conjurapi.LdapMappingListResponse, error)
	deleteUserMapping  func(*testing.T, string, string) error
	clientFactoryError error
	assert             func(*testing.T, string, string, error)
}

var ldapUserCmdTestCases = []ldapUserCmdTestCase{
	{
		name: "users create success table",
		args: []string{"authn-ldap", "--id", "ad-prod", "users", "create", "jan.k", "--roles", "special-role"},
		createUserMapping: func(t *testing.T, serviceID, username string, roles []string) (*conjurapi.LdapUserMappingResponse, error) {
			assert.Equal(t, "ad-prod", serviceID)
			assert.Equal(t, "jan.k", username)
			assert.Equal(t, []string{"special-role"}, roles)
			return &conjurapi.LdapUserMappingResponse{LdapUsername: "jan.k", Roles: []string{"special-role"}}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.NoError(t, err)
			assert.Empty(t, stderr)
			assert.Contains(t, stdout, "jan.k")
			assert.Contains(t, stdout, "special-role")
		},
	},
	{
		name: "users create success json",
		args: []string{"authn-ldap", "--id", "ad-prod", "--output", "json", "users", "create", "jan.k", "--roles", "special-role"},
		createUserMapping: func(t *testing.T, serviceID, username string, roles []string) (*conjurapi.LdapUserMappingResponse, error) {
			return &conjurapi.LdapUserMappingResponse{LdapUsername: "jan.k", Roles: []string{"special-role"}}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.NoError(t, err)
			var result conjurapi.LdapUserMappingResponse
			assert.NoError(t, json.Unmarshal([]byte(stdout), &result))
			assert.Equal(t, "jan.k", result.LdapUsername)
			assert.Equal(t, []string{"special-role"}, result.Roles)
		},
	},
	{
		name: "users create API error",
		args: []string{"authn-ldap", "--id", "ad-prod", "users", "create", "jan.k", "--roles", "special-role"},
		createUserMapping: func(t *testing.T, serviceID, username string, roles []string) (*conjurapi.LdapUserMappingResponse, error) {
			return nil, errors.New("404 Not Found")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to create user mapping 'jan.k'")
			assert.Contains(t, err.Error(), "404 Not Found")
		},
	},
	{
		name:               "users create client factory error",
		args:               []string{"authn-ldap", "--id", "ad-prod", "users", "create", "jan.k", "--roles", "special-role"},
		clientFactoryError: errors.New("client factory error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, stderr, "Error: client factory error")
		},
	},
	{
		name: "users show success",
		args: []string{"authn-ldap", "--id", "ad-prod", "users", "show", "jan.k"},
		showUserMapping: func(t *testing.T, serviceID, username string) (*conjurapi.LdapUserMappingResponse, error) {
			assert.Equal(t, "ad-prod", serviceID)
			assert.Equal(t, "jan.k", username)
			return &conjurapi.LdapUserMappingResponse{LdapUsername: "jan.k", Roles: []string{"special-role"}}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.NoError(t, err)
			assert.Contains(t, stdout, "jan.k")
			assert.Contains(t, stdout, "special-role")
		},
	},
	{
		name: "users show API error",
		args: []string{"authn-ldap", "--id", "ad-prod", "users", "show", "jan.k"},
		showUserMapping: func(t *testing.T, serviceID, username string) (*conjurapi.LdapUserMappingResponse, error) {
			return nil, errors.New("404 Not Found")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to show user mapping 'jan.k'")
		},
	},
	{
		name: "users list success table",
		args: []string{"authn-ldap", "--id", "ad-prod", "users", "list"},
		listUserMappings: func(t *testing.T, serviceID string) (*conjurapi.LdapMappingListResponse, error) {
			assert.Equal(t, "ad-prod", serviceID)
			return &conjurapi.LdapMappingListResponse{Keys: []string{"jan.k", "admin.user"}}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.NoError(t, err)
			assert.Contains(t, stdout, "jan.k")
			assert.Contains(t, stdout, "admin.user")
		},
	},
	{
		name: "users list success json",
		args: []string{"authn-ldap", "--id", "ad-prod", "--output", "json", "users", "list"},
		listUserMappings: func(t *testing.T, serviceID string) (*conjurapi.LdapMappingListResponse, error) {
			return &conjurapi.LdapMappingListResponse{Keys: []string{"jan.k", "admin.user"}}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.NoError(t, err)
			var result conjurapi.LdapMappingListResponse
			assert.NoError(t, json.Unmarshal([]byte(stdout), &result))
			assert.Equal(t, []string{"jan.k", "admin.user"}, result.Keys)
		},
	},
	{
		name: "users list API error",
		args: []string{"authn-ldap", "--id", "ad-prod", "users", "list"},
		listUserMappings: func(t *testing.T, serviceID string) (*conjurapi.LdapMappingListResponse, error) {
			return nil, errors.New("403 Forbidden")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to list user mappings")
		},
	},
	{
		name: "users delete success",
		args: []string{"authn-ldap", "--id", "ad-prod", "users", "delete", "jan.k"},
		deleteUserMapping: func(t *testing.T, serviceID, username string) error {
			assert.Equal(t, "ad-prod", serviceID)
			assert.Equal(t, "jan.k", username)
			return nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.NoError(t, err)
			assert.Contains(t, stdout, "Success! LDAP user mapping 'jan.k' deleted")
		},
	},
	{
		name: "users delete API error",
		args: []string{"authn-ldap", "--id", "ad-prod", "users", "delete", "jan.k"},
		deleteUserMapping: func(t *testing.T, serviceID, username string) error {
			return errors.New("404 Not Found")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to delete user mapping 'jan.k'")
		},
	},
	{
		name: "users list invalid output format",
		args: []string{"authn-ldap", "--id", "ad-prod", "--output", "xml", "users", "list"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid output format")
		},
	},
	{
		name: "users list missing id flag",
		args: []string{"authn-ldap", "users", "list"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, stderr, "required flag(s)")
		},
	},
}

func TestLdapUsersCmd(t *testing.T) {
	t.Parallel()
	for _, tc := range ldapUserCmdTestCases {
		t.Run(tc.name, func(t *testing.T) {
			mockGroup := mockLdapGroupClient{t: t}
			mockUser := mockLdapUserClient{
				t:                 t,
				createUserMapping: tc.createUserMapping,
				showUserMapping:   tc.showUserMapping,
				listUserMappings:  tc.listUserMappings,
				deleteUserMapping: tc.deleteUserMapping,
			}

			cmd := newAuthnLdapCommand(
				func(_ *cobra.Command) (ldapGroupMappingClient, error) {
					return mockGroup, nil
				},
				func(_ *cobra.Command) (ldapUserMappingClient, error) {
					return mockUser, tc.clientFactoryError
				},
			)

			stdout, stderr, err := executeCommandForTest(t, cmd, tc.args...)
			tc.assert(t, stdout, stderr, err)
		})
	}
}
