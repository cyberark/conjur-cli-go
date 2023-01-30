package cmd

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type mockResourceClient struct {
	t              *testing.T
	resourceExists func(t *testing.T, resourceID string) (bool, error)
	resource       func(t *testing.T, resourceID string) (resource map[string]interface{}, err error)
	permittedRoles func(t *testing.T, resourceID, privilege string) ([]string, error)
}

func (m mockResourceClient) ResourceExists(resourceID string) (bool, error) {
	return m.resourceExists(m.t, resourceID)
}

func (m mockResourceClient) Resource(resourceID string) (resource map[string]interface{}, err error) {
	return m.resource(m.t, resourceID)
}

func (m mockResourceClient) PermittedRoles(resourceID, privilege string) ([]string, error) {
	return m.permittedRoles(m.t, resourceID, privilege)
}

type resourceCmdTestCase struct {
	name               string
	args               []string
	resourceExists     func(t *testing.T, resourceID string) (bool, error)
	resource           func(t *testing.T, resourceID string) (resource map[string]interface{}, err error)
	permittedRoles     func(t *testing.T, resourceID string, privilege string) ([]string, error)
	clientFactoryError error
	assert             func(t *testing.T, stdout string, stderr string, err error)
}

var resourceExistsCmdTestCases = []resourceCmdTestCase{
	{
		name: "resource exists help",
		args: []string{"exists", "--help"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "resource exists missing resource-id",
		args: []string{"exists"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "resource exists missing resource-id with json flag",
		args: []string{"exists", "--json"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "resource exists return true",
		args: []string{"exists", "meow"},
		resourceExists: func(t *testing.T, resourceID string) (bool, error) {
			assert.Equal(t, "meow", resourceID)

			return true, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "true")
		},
	},
	{
		name: "resource exists return false",
		args: []string{"exists", "meow"},
		resourceExists: func(t *testing.T, resourceID string) (bool, error) {
			assert.Equal(t, "meow", resourceID)

			return false, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "false")
		},
	},
	{
		name: "resource exists returns json true",
		args: []string{"exists", "--json", "meow"},
		resourceExists: func(t *testing.T, roleID string) (bool, error) {
			assert.Equal(t, "meow", roleID)

			return true, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "{\n  \"exists\": true\n}\n")
		},
	},
	{
		name: "resource exists returns json false",
		args: []string{"exists", "--json", "meow"},
		resourceExists: func(t *testing.T, roleID string) (bool, error) {
			assert.Equal(t, "meow", roleID)

			return false, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "{\n  \"exists\": false\n}\n")
		},
	},
	{
		name: "resource exists client error",
		args: []string{"exists", "abcdefg"},
		resourceExists: func(t *testing.T, resourceID string) (bool, error) {
			return false, fmt.Errorf("%s", "an error")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: an error\n")
		},
	},
	{
		name:               "resource exists client factory error",
		args:               []string{"exists", "abcdefg"},
		clientFactoryError: fmt.Errorf("%s", "client factory error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: client factory error\n")
		},
	},
}

var resourcePermittedRolesCmdTestCases = []resourceCmdTestCase{
	{
		name: "resource permitted-roles help",
		args: []string{"permitted-roles", "--help"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "resource permitted-roles missing resource-id",
		args: []string{"permitted-roles", "execute"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "resource permitted-roles missing privilege",
		args: []string{"permitted-roles", "meow"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "resource permitted-roles return roles found in pretty format",
		args: []string{"permitted-roles", "meow", "moo"},
		permittedRoles: func(t *testing.T, resourceID, privilege string) ([]string, error) {
			assert.Equal(t, "meow", resourceID)
			assert.Equal(t, "moo", privilege)

			return []string{"role1", "role2"}, nil
		},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "[\n  \"role1\",\n  \"role2\"\n]\n")
		},
	},
	{
		name: "resource permitted-roles return no roles found",
		args: []string{"permitted-roles", "meow", "moo"},
		permittedRoles: func(t *testing.T, resourceID, privilege string) ([]string, error) {
			assert.Equal(t, "meow", resourceID)
			assert.Equal(t, "moo", privilege)

			return nil, fmt.Errorf("%s", "an error")
		},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stderr, "Error: an error\n")
		},
	},
	{
		name: "resource permitted-roles client error",
		args: []string{"permitted-roles", "abcdef", "hijklm"},
		permittedRoles: func(t *testing.T, resourceID, privilege string) ([]string, error) {
			return nil, fmt.Errorf("%s", "an error")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: an error\n")
		},
	},
	{
		name:               "resource permitted-roles client factory error",
		args:               []string{"permitted-roles", "abcdefg", "hijklm"},
		clientFactoryError: fmt.Errorf("%s", "client factory error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: client factory error\n")
		},
	},
}

var resourceShowCmdTestCases = []resourceCmdTestCase{
	{
		name: "resource show help",
		args: []string{"show", "--help"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "resource show missing resource-id",
		args: []string{"show"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "resource show return object in pretty format",
		args: []string{"show", "meow"},
		resource: func(t *testing.T, resourceID string) (resource map[string]interface{}, err error) {
			assert.Equal(t, "meow", resourceID)

			return map[string]interface{}{"key": "value"}, nil
		},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "{\n  \"key\": \"value\"\n}\n")
		},
	},
	{
		name: "resource show return object not found",
		args: []string{"show", "meow"},
		resource: func(t *testing.T, resourceID string) (resource map[string]interface{}, err error) {
			assert.Equal(t, "meow", resourceID)

			return nil, fmt.Errorf("%s", "an error")
		},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stderr, "Error: an error\n")
		},
	},
	{
		name: "resource show client error",
		args: []string{"show", "abcdefg"},
		resource: func(t *testing.T, resourceID string) (resource map[string]interface{}, err error) {
			return nil, fmt.Errorf("%s", "an error")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: an error\n")
		},
	},
	{
		name:               "resource show client factory error",
		args:               []string{"show", "abcdefg"},
		clientFactoryError: fmt.Errorf("%s", "client factory error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: client factory error\n")
		},
	},
}

func TestResourceExistsCmd(t *testing.T) {
	for _, tc := range resourceExistsCmdTestCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := mockResourceClient{t: t, resourceExists: tc.resourceExists}

			cmd := newResourceExistsCmd(
				func(cmd *cobra.Command) (resourceClient, error) {
					return mockClient, tc.clientFactoryError
				},
			)

			stdout, stderr, err := executeCommandForTest(t, cmd, tc.args...)
			tc.assert(t, stdout, stderr, err)
		})
	}
}

func TestResourcePermittedRolesCmd(t *testing.T) {
	for _, tc := range resourcePermittedRolesCmdTestCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := mockResourceClient{t: t, permittedRoles: tc.permittedRoles}

			cmd := newResourcePermittedRolesCmd(
				func(cmd *cobra.Command) (resourceClient, error) {
					return mockClient, tc.clientFactoryError
				},
			)

			stdout, stderr, err := executeCommandForTest(t, cmd, tc.args...)
			tc.assert(t, stdout, stderr, err)
		})
	}
}

func TestResourceShowCmd(t *testing.T) {
	for _, tc := range resourceShowCmdTestCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := mockResourceClient{t: t, resource: tc.resource}

			cmd := newResourceShowCmd(
				func(cmd *cobra.Command) (resourceClient, error) {
					return mockClient, tc.clientFactoryError
				},
			)

			stdout, stderr, err := executeCommandForTest(t, cmd, tc.args...)
			tc.assert(t, stdout, stderr, err)
		})
	}
}
