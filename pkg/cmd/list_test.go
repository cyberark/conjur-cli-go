package cmd

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

var clientResponseStr = `[
  {
    "annotations": [],
    "created_at": "2023-02-07T20:37:23.400+00:00",
    "id": "dev:host:test-host",
    "owner": "dev:user:admin",
    "permissions": [],
    "policy": "dev:policy:root",
    "restricted_to": []
  },
  {
    "annotations": [],
    "created_at": "2023-02-07T20:37:23.400+00:00",
    "id": "dev:layer:test-layer",
    "owner": "dev:user:admin",
    "permissions": [],
    "policy": "dev:policy:root"
  }
]
`

var clientResponseIDsStr = `[
  "dev:host:test-host",
  "dev:layer:test-layer"
]
`

var clientResponse = make([]map[string]interface{}, 1)
var _ = json.Unmarshal([]byte(clientResponseStr), &clientResponse)

type mockListClient struct {
	t              *testing.T
	listResources  func(t *testing.T, filter *conjurapi.ResourceFilter) ([]map[string]interface{}, error)
	resourcesCount func(t *testing.T, filter *conjurapi.ResourceFilter) (*conjurapi.ResourcesCount, error)
}

func (m mockListClient) Resources(filter *conjurapi.ResourceFilter) ([]map[string]interface{}, error) {
	return m.listResources(m.t, filter)
}

func (m mockListClient) ResourcesCount(filter *conjurapi.ResourceFilter) (*conjurapi.ResourcesCount, error) {
	return m.resourcesCount(m.t, filter)
}

var listCmdTestCases = []struct {
	name               string
	args               []string
	listResources      func(t *testing.T, filter *conjurapi.ResourceFilter) ([]map[string]interface{}, error)
	resourcesCount     func(t *testing.T, filter *conjurapi.ResourceFilter) (*conjurapi.ResourcesCount, error)
	roleMembers        func(t *testing.T, roleID string) (members []map[string]interface{}, err error)
	permittedRoles     func(t *testing.T, resourceID, privilege string) ([]string, error)
	clientFactoryError error
	assert             func(t *testing.T, stdout string, stderr string, err error)
}{
	{
		name: "list help",
		args: []string{"list", "--help"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "list kind",
		args: []string{"list", "-k", "woof"},
		listResources: func(t *testing.T, filter *conjurapi.ResourceFilter) ([]map[string]interface{}, error) {
			assert.Equal(t, "woof", filter.Kind)

			return clientResponse, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "test-host")
		},
	},
	{
		name: "list search",
		args: []string{"list", "-s", "friend"},
		listResources: func(t *testing.T, filter *conjurapi.ResourceFilter) ([]map[string]interface{}, error) {
			assert.Equal(t, "friend", filter.Search)

			return clientResponse, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "test-host")
		},
	},
	{
		name: "list limit",
		args: []string{"list", "-l", "5"},
		listResources: func(t *testing.T, filter *conjurapi.ResourceFilter) ([]map[string]interface{}, error) {
			assert.Equal(t, 5, filter.Limit)

			return clientResponse, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "test-host")
		},
	},
	{
		name: "list offset",
		args: []string{"list", "-o", "2"},
		listResources: func(t *testing.T, filter *conjurapi.ResourceFilter) ([]map[string]interface{}, error) {
			assert.Equal(t, 2, filter.Offset)

			return clientResponse, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "test-host")
		},
	},
	{
		name: "list role",
		args: []string{"list", "-r", "dev:host:test-host"},
		listResources: func(t *testing.T, filter *conjurapi.ResourceFilter) ([]map[string]interface{}, error) {
			assert.Equal(t, "dev:host:test-host", filter.Role)

			return clientResponse, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "test-host")
		},
	},
	{
		name: "list inspect",
		args: []string{"list", "--inspect"},
		listResources: func(t *testing.T, filter *conjurapi.ResourceFilter) ([]map[string]interface{}, error) {
			return clientResponse, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Equal(t, stdout, clientResponseStr)
		},
	},
	// BEGIN COMPATIBILITY WITH PYTHON CLI
	{
		name: "list members",
		args: []string{"list", "--members-of", "dev:user:alice"},
		roleMembers: func(t *testing.T, roleID string) (members []map[string]interface{}, err error) {
			assert.Equal(t, "dev:user:alice", roleID)

			return []map[string]interface{}{{"admin_option": "true", "role": "dev:user:alice", "member": "dev:user:bob"}}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "Flag --members-of has been deprecated, Use 'role members' instead")
			assert.Contains(t, stdout, `[
  "dev:user:bob"
]
`)
		},
	},
	{
		name: "list permitted-roles",
		args: []string{"list", "--permitted-roles", "dev:user:alice", "--privilege", "read"},
		permittedRoles: func(t *testing.T, resourceID, privilege string) ([]string, error) {
			assert.Equal(t, "dev:user:alice", resourceID)
			assert.Equal(t, "read", privilege)

			return []string{"dev:user:bob"}, nil
		},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "Flag --permitted-roles has been deprecated, Use 'resource permitted-roles' instead")
			assert.Contains(t, stdout, "Flag --privilege has been deprecated, Use 'resource permitted-roles' instead")
			assert.Contains(t, stdout, `[
  "dev:user:bob"
]
`)
		},
	},
	// END COMPATIBILITY WITH PYTHON CLI
	{
		name: "list no flags",
		args: []string{"list"},
		listResources: func(t *testing.T, filter *conjurapi.ResourceFilter) ([]map[string]interface{}, error) {
			assert.Equal(t, "", filter.Kind)
			assert.Equal(t, "", filter.Search)
			assert.Equal(t, 0, filter.Limit)
			assert.Equal(t, 0, filter.Offset)

			return clientResponse, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "test-host")
		},
	},
	{
		name: "list all flags",
		args: []string{"list", "-k", "qwerty", "-s", "abc", "-l", "5", "-o", "2"},
		listResources: func(t *testing.T, filter *conjurapi.ResourceFilter) ([]map[string]interface{}, error) {
			assert.Equal(t, "qwerty", filter.Kind)
			assert.Equal(t, "abc", filter.Search)
			assert.Equal(t, 5, filter.Limit)
			assert.Equal(t, 2, filter.Offset)

			return clientResponse, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "test-host")
		},
	},
	{
		name: "list resource IDs",
		args: []string{"list"},
		listResources: func(t *testing.T, filter *conjurapi.ResourceFilter) ([]map[string]interface{}, error) {
			return clientResponse, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Equal(t, stdout, clientResponseIDsStr)
		},
	},
	{
		name: "list client error",
		args: []string{"list", "-k", "asdf", "-s", "qwer"},
		listResources: func(t *testing.T, filter *conjurapi.ResourceFilter) ([]map[string]interface{}, error) {
			return nil, fmt.Errorf("%s", "an error")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: an error\n")
		},
	},
	{
		name:               "list client factory error",
		args:               []string{"list", "-k", "asdf", "-s", "qwer"},
		clientFactoryError: fmt.Errorf("%s", "client factory error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: client factory error\n")
		},
	},
	{
		name: "list count",
		args: []string{"list", "-c"},
		resourcesCount: func(t *testing.T, filter *conjurapi.ResourceFilter) (*conjurapi.ResourcesCount, error) {
			return &conjurapi.ResourcesCount{Count: 3}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, `"count": 3`)
		},
	},
	{
		name: "list count client error",
		args: []string{"list", "-c"},
		resourcesCount: func(t *testing.T, filter *conjurapi.ResourceFilter) (*conjurapi.ResourcesCount, error) {
			return nil, fmt.Errorf("%s", "an error")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: an error\n")
		},
	},
}

func TestListCmd(t *testing.T) {
	for _, tc := range listCmdTestCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := mockListClient{t: t, listResources: tc.listResources, resourcesCount: tc.resourcesCount}
			mockRoleClient := mockRoleClient{t: t, roleMembers: tc.roleMembers}
			mockResourceClient := mockResourceClient{t: t, permittedRoles: tc.permittedRoles}

			cmd := newListCmd(
				func(cmd *cobra.Command) (listClient, error) {
					return mockClient, tc.clientFactoryError
				},
				func(cmd *cobra.Command) (roleClient, error) {
					return mockRoleClient, tc.clientFactoryError
				},
				func(cmd *cobra.Command) (resourceClient, error) {
					return mockResourceClient, tc.clientFactoryError
				},
			)

			stdout, stderr, err := executeCommandForTest(t, cmd, tc.args...)
			tc.assert(t, stdout, stderr, err)
		})
	}
}
