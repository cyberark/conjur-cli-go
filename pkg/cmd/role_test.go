package cmd

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type mockRoleClient struct {
	t                  *testing.T
	roleExists         func(t *testing.T, roleID string) (bool, error)
	role               func(t *testing.T, roleID string) (role map[string]interface{}, err error)
	roleMembers        func(t *testing.T, roleID string) (members []map[string]interface{}, err error)
	roleMembershipsAll func(t *testing.T, roleID string) (memberships []string, err error)
}

func (m mockRoleClient) RoleExists(roleID string) (bool, error) {
	return m.roleExists(m.t, roleID)
}

func (m mockRoleClient) Role(roleID string) (role map[string]interface{}, err error) {
	return m.role(m.t, roleID)
}

func (m mockRoleClient) RoleMembers(roleID string) (members []map[string]interface{}, err error) {
	return m.roleMembers(m.t, roleID)
}

func (m mockRoleClient) RoleMembershipsAll(roleID string) (memberships []string, err error) {
	return m.roleMembershipsAll(m.t, roleID)
}

type roleCmdTestCase struct {
	name               string
	args               []string
	roleExists         func(t *testing.T, roleID string) (bool, error)
	role               func(t *testing.T, roleID string) (role map[string]interface{}, err error)
	roleMembers        func(t *testing.T, roleID string) (members []map[string]interface{}, err error)
	roleMembershipsAll func(t *testing.T, roleID string) (memberships []string, err error)
	clientFactoryError error
	assert             func(t *testing.T, stdout string, stderr string, err error)
}

var roleExistsCmdTestCases = []roleCmdTestCase{
	{
		name: "role exists help",
		args: []string{"exists", "--help"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "role exists missing role-id",
		args: []string{"exists"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "role exists missing role-id with json flag",
		args: []string{"exists", "--json"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "role exists return true",
		args: []string{"exists", "meow"},

		roleExists: func(t *testing.T, roleID string) (bool, error) {
			assert.Equal(t, "meow", roleID)

			return true, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "true")
		},
	},
	{
		name: "role exists return false",
		args: []string{"exists", "meow"},
		roleExists: func(t *testing.T, roleID string) (bool, error) {
			assert.Equal(t, "meow", roleID)

			return false, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "false")
		},
	},
	{
		name: "role exists returns json true",
		args: []string{"exists", "--json", "meow"},

		roleExists: func(t *testing.T, roleID string) (bool, error) {
			assert.Equal(t, "meow", roleID)

			return true, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "{\n  \"exists\": true\n}\n")
		},
	},
	{
		name: "role exists returns json false",
		args: []string{"exists", "--json", "meow"},

		roleExists: func(t *testing.T, roleID string) (bool, error) {
			assert.Equal(t, "meow", roleID)

			return false, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "{\n  \"exists\": false\n}\n")
		},
	},
	{
		name: "role exists return true",
		args: []string{"exists", "meow"},

		roleExists: func(t *testing.T, roleID string) (bool, error) {
			assert.Equal(t, "meow", roleID)

			return true, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "true")
		},
	},
	{
		name: "role exists client error",
		args: []string{"exists", "abcdefg"},
		roleExists: func(t *testing.T, roleID string) (bool, error) {
			return false, fmt.Errorf("%s", "an error")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: an error\n")
		},
	},
	{
		name:               "role exists client factory error",
		args:               []string{"exists", "abcdefg"},
		clientFactoryError: fmt.Errorf("%s", "client factory error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: client factory error\n")
		},
	},
}

var roleShowCmdTestCases = []roleCmdTestCase{
	{
		name: "role show help",
		args: []string{"show", "--help"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "role show missing role-id",
		args: []string{"show"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "role show return object in pretty format",
		args: []string{"show", "meow"},
		role: func(t *testing.T, roleID string) (role map[string]interface{}, err error) {
			assert.Equal(t, "meow", roleID)

			return map[string]interface{}{"key": "value"}, nil
		},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "{\n  \"key\": \"value\"\n}\n")
		},
	},
	{
		name: "role show return object not found",
		args: []string{"show", "meow"},
		role: func(t *testing.T, roleID string) (role map[string]interface{}, err error) {
			assert.Equal(t, "meow", roleID)

			return nil, fmt.Errorf("%s", "an error")
		},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stderr, "Error: an error\n")
		},
	},
	{
		name: "role show client error",
		args: []string{"show", "abcdefg"},
		role: func(t *testing.T, roleID string) (role map[string]interface{}, err error) {
			return nil, fmt.Errorf("%s", "an error")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: an error\n")
		},
	},
	{
		name:               "role show client factory error",
		args:               []string{"show", "abcdefg"},
		clientFactoryError: fmt.Errorf("%s", "client factory error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: client factory error\n")
		},
	},
}

var roleMembersCmdTestCases = []roleCmdTestCase{
	{
		name: "role members help",
		args: []string{"members", "--help"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "role members missing role-id",
		args: []string{"members"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "role members missing role-id with verbose flag",
		args: []string{"members", "--verbose"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "role members return verbose members in pretty format",
		args: []string{"members", "-v", "meow"},
		roleMembers: func(t *testing.T, roleID string) (members []map[string]interface{}, err error) {
			assert.Equal(t, "meow", roleID)

			return []map[string]interface{}{{"admin_option": "true", "role": "abc", "member": "def"}}, nil
		},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "[\n  {\n    \"admin_option\": \"true\",\n    \"member\": \"def\",\n    \"role\": \"abc\"\n  }\n]\n")
		},
	},
	{
		name: "role members return members in pretty format",
		args: []string{"members", "meow"},
		roleMembers: func(t *testing.T, roleID string) (members []map[string]interface{}, err error) {
			assert.Equal(t, "meow", roleID)

			return []map[string]interface{}{{"admin_option": "true", "role": "abc", "member": "member1"}}, nil
		},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "[\n  \"member1\"\n]\n")
		},
	},
	{
		name: "role members return no members",
		args: []string{"members", "meow"},
		roleMembers: func(t *testing.T, roleID string) (members []map[string]interface{}, err error) {
			assert.Equal(t, "meow", roleID)

			return nil, fmt.Errorf("%s", "an error")
		},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stderr, "Error: an error\n")
		},
	},
	{
		name: "role members client error",
		args: []string{"members", "abcdefg"},
		roleMembers: func(t *testing.T, roleID string) (members []map[string]interface{}, err error) {
			return nil, fmt.Errorf("%s", "an error")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: an error\n")
		},
	},
	{
		name:               "role members client factory error",
		args:               []string{"members", "abcdefg"},
		clientFactoryError: fmt.Errorf("%s", "client factory error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: client factory error\n")
		},
	},
}

var roleMembershipsCmdTestCases = []roleCmdTestCase{
	{
		name: "role memberships help",
		args: []string{"memberships", "--help"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "role memberships missing role-id",
		args: []string{"memberships"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "role memberships return memberships in pretty format",
		args: []string{"memberships", "meow"},
		roleMembershipsAll: func(t *testing.T, roleID string) (memberships []string, err error) {
			assert.Equal(t, "meow", roleID)

			return []string{"role1"}, nil
		},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "[\n  \"role1\"\n]\n")
		},
	},
	{
		name: "role memberships return no members",
		args: []string{"memberships", "meow"},
		roleMembershipsAll: func(t *testing.T, roleID string) (memberships []string, err error) {
			assert.Equal(t, "meow", roleID)

			return nil, fmt.Errorf("%s", "an error")
		},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stderr, "Error: an error\n")
		},
	},
	{
		name: "role memberships client error",
		args: []string{"memberships", "abcdefg"},
		roleMembershipsAll: func(t *testing.T, roleID string) (memberships []string, err error) {
			return nil, fmt.Errorf("%s", "an error")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: an error\n")
		},
	},
	{
		name:               "role memberships client factory error",
		args:               []string{"memberships", "abcdefg"},
		clientFactoryError: fmt.Errorf("%s", "client factory error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: client factory error\n")
		},
	},
}

func TestRoleExistsCmd(t *testing.T) {
	for _, tc := range roleExistsCmdTestCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := mockRoleClient{t: t, roleExists: tc.roleExists}

			cmd := newRoleExistsCmd(
				func(cmd *cobra.Command) (roleClient, error) {
					return mockClient, tc.clientFactoryError
				},
			)

			stdout, stderr, err := executeCommandForTest(t, cmd, tc.args...)
			tc.assert(t, stdout, stderr, err)
		})
	}
}

func TestRoleShowCmd(t *testing.T) {
	for _, tc := range roleShowCmdTestCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := mockRoleClient{t: t, role: tc.role}

			cmd := newRoleShowCmd(
				func(cmd *cobra.Command) (roleClient, error) {
					return mockClient, tc.clientFactoryError
				},
			)

			stdout, stderr, err := executeCommandForTest(t, cmd, tc.args...)
			tc.assert(t, stdout, stderr, err)
		})
	}
}

func TestRoleMembersCmd(t *testing.T) {
	for _, tc := range roleMembersCmdTestCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := mockRoleClient{t: t, roleMembers: tc.roleMembers}

			cmd := newRoleMembersCmd(
				func(cmd *cobra.Command) (roleClient, error) {
					return mockClient, tc.clientFactoryError
				},
			)

			stdout, stderr, err := executeCommandForTest(t, cmd, tc.args...)
			tc.assert(t, stdout, stderr, err)
		})
	}
}

func TestRoleMembershipsCmd(t *testing.T) {
	for _, tc := range roleMembershipsCmdTestCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := mockRoleClient{t: t, roleMembershipsAll: tc.roleMembershipsAll}

			cmd := newRoleMembershipsCmd(
				func(cmd *cobra.Command) (roleClient, error) {
					return mockClient, tc.clientFactoryError
				},
			)

			stdout, stderr, err := executeCommandForTest(t, cmd, tc.args...)
			tc.assert(t, stdout, stderr, err)
		})
	}
}
