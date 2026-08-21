package cmd

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-cli-go/pkg/pagination"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			assert.Contains(t, stdout, "[\n  \"dev:user:bob\"\n]\n")
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
			assert.Contains(t, stdout, "[\n  \"dev:user:bob\"\n]\n")
		},
	},
	// END COMPATIBILITY WITH PYTHON CLI
	{
		name: "list no flags",
		args: []string{"list"},
		listResources: func(t *testing.T, filter *conjurapi.ResourceFilter) ([]map[string]interface{}, error) {
			assert.Equal(t, "", filter.Kind)
			assert.Equal(t, "", filter.Search)
			// No explicit --limit: the first page is fetched at the
			// server's per-request page size, not the raw flag default of 0.
			assert.Equal(t, pagination.EffectivePageSize(), filter.Limit)
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
	{
		name: "list count explicit output json",
		args: []string{"list", "-c", "--output", "json"},
		resourcesCount: func(t *testing.T, filter *conjurapi.ResourceFilter) (*conjurapi.ResourcesCount, error) {
			return &conjurapi.ResourcesCount{Count: 3}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, `"count": 3`)
		},
	},
	{
		name: "list inspect explicit output json",
		args: []string{"list", "--inspect", "--output", "json"},
		listResources: func(t *testing.T, filter *conjurapi.ResourceFilter) ([]map[string]interface{}, error) {
			return clientResponse, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Equal(t, stdout, clientResponseStr)
		},
	},
	{
		name: "list default explicit output json",
		args: []string{"list", "--output", "json"},
		listResources: func(t *testing.T, filter *conjurapi.ResourceFilter) ([]map[string]interface{}, error) {
			return clientResponse, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Equal(t, stdout, clientResponseIDsStr)
		},
	},
	{
		name: "list inspect explicit custom-columns",
		args: []string{"list", "--inspect", "--output", "custom-columns=ID:.id"},
		listResources: func(t *testing.T, filter *conjurapi.ResourceFilter) ([]map[string]interface{}, error) {
			return []map[string]interface{}{{"id": "dev:variable:foo"}}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Equal(t, "ID\ndev:variable:foo\n", stdout)
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

// makeResourceItems returns count dummy resource maps labeled by their
// absolute position, starting at offset.
func makeResourceItems(offset, count int) []map[string]interface{} {
	items := make([]map[string]interface{}, count)
	for i := 0; i < count; i++ {
		items[i] = map[string]interface{}{"id": fmt.Sprintf("dev:host:item-%d", offset+i)}
	}
	return items
}

// datasetListResources returns a listResources func backed by a fixed-size
// in-memory dataset of `total` items, plus a pointer to the number of
// calls made (for assertions on request count).
func datasetListResources(total int) (func(t *testing.T, filter *conjurapi.ResourceFilter) ([]map[string]interface{}, error), *int) {
	calls := 0
	var mu sync.Mutex
	fetch := func(t *testing.T, filter *conjurapi.ResourceFilter) ([]map[string]interface{}, error) {
		mu.Lock()
		calls++
		mu.Unlock()
		offset := filter.Offset
		limit := filter.Limit
		if offset >= total {
			return nil, nil
		}
		end := offset + limit
		if end > total {
			end = total
		}
		return makeResourceItems(offset, end-offset), nil
	}
	return fetch, &calls
}

func TestListCmd_Pagination(t *testing.T) {
	t.Run("auto-paginates above the server page size", func(t *testing.T) {
		t.Setenv("CONJUR_LIST_PAGE_SIZE", "2")
		listResources, calls := datasetListResources(5)
		mockClient := mockListClient{t: t, listResources: listResources}

		cmd := newListCmd(
			func(cmd *cobra.Command) (listClient, error) { return mockClient, nil },
			func(cmd *cobra.Command) (roleClient, error) { return nil, nil },
			func(cmd *cobra.Command) (resourceClient, error) { return nil, nil },
		)

		stdout, _, err := executeCommandForTest(t, cmd, "list", "--limit=-1", "-i")
		require.NoError(t, err)

		var got []map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(stdout), &got))
		assert.Len(t, got, 5)
		assert.Equal(t, "dev:host:item-0", got[0]["id"])
		assert.Equal(t, "dev:host:item-4", got[4]["id"])
		assert.GreaterOrEqual(t, *calls, 3, "expected multiple page requests (pages of 2 for 5 items)")
	})

	t.Run("warns to stderr when the implicit default cap truncates results", func(t *testing.T) {
		t.Setenv("CONJUR_LIST_PAGE_SIZE", "2")
		listResources, _ := datasetListResources(1_000_000) // effectively unbounded
		mockClient := mockListClient{t: t, listResources: listResources}

		cmd := newListCmd(
			func(cmd *cobra.Command) (listClient, error) { return mockClient, nil },
			func(cmd *cobra.Command) (roleClient, error) { return nil, nil },
			func(cmd *cobra.Command) (resourceClient, error) { return nil, nil },
		)

		// No --limit passed: implicit cap is 10 * PageSize = 20 with
		// PageSize overridden to 2 above.
		_, stderr, err := executeCommandForTest(t, cmd, "list")
		require.NoError(t, err)
		assert.Contains(t, stderr, "may be truncated at 20 items")
		assert.Contains(t, stderr, "--limit -1")
	})

	t.Run("--offset composes correctly with multi-page pagination", func(t *testing.T) {
		t.Setenv("CONJUR_LIST_PAGE_SIZE", "2")
		listResources, _ := datasetListResources(10)
		mockClient := mockListClient{t: t, listResources: listResources}

		cmd := newListCmd(
			func(cmd *cobra.Command) (listClient, error) { return mockClient, nil },
			func(cmd *cobra.Command) (roleClient, error) { return nil, nil },
			func(cmd *cobra.Command) (resourceClient, error) { return nil, nil },
		)

		stdout, _, err := executeCommandForTest(t, cmd, "list", "--offset", "3", "--limit=-1", "-i")
		require.NoError(t, err)

		var got []map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(stdout), &got))
		require.Len(t, got, 7, "10 items total, starting at offset 3, leaves 7")
		assert.Equal(t, "dev:host:item-3", got[0]["id"], "first item should start at the absolute offset, not offset 0")
		assert.Equal(t, "dev:host:item-9", got[6]["id"])
	})

	t.Run("explicit --limit that returns exactly N items does not warn", func(t *testing.T) {
		listResources, _ := datasetListResources(5)
		mockClient := mockListClient{t: t, listResources: listResources}

		cmd := newListCmd(
			func(cmd *cobra.Command) (listClient, error) { return mockClient, nil },
			func(cmd *cobra.Command) (roleClient, error) { return nil, nil },
			func(cmd *cobra.Command) (resourceClient, error) { return nil, nil },
		)

		stdout, stderr, err := executeCommandForTest(t, cmd, "list", "--limit", "5", "-i")
		require.NoError(t, err)

		var got []map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(stdout), &got))
		assert.Len(t, got, 5)
		assert.NotContains(t, stderr, "may be truncated")
		assert.Empty(t, stderr)
	})

	t.Run("--limit 0 makes no API call and returns nothing", func(t *testing.T) {
		listResources, calls := datasetListResources(100)
		mockClient := mockListClient{t: t, listResources: listResources}

		cmd := newListCmd(
			func(cmd *cobra.Command) (listClient, error) { return mockClient, nil },
			func(cmd *cobra.Command) (roleClient, error) { return nil, nil },
			func(cmd *cobra.Command) (resourceClient, error) { return nil, nil },
		)

		stdout, _, err := executeCommandForTest(t, cmd, "list", "-l", "0")
		require.NoError(t, err)

		var got []string
		require.NoError(t, json.Unmarshal([]byte(stdout), &got))
		assert.Empty(t, got)
		assert.Equal(t, 0, *calls)
	})

	t.Run("rejects --limit values below -1", func(t *testing.T) {
		mockClient := mockListClient{t: t}
		cmd := newListCmd(
			func(cmd *cobra.Command) (listClient, error) { return mockClient, nil },
			func(cmd *cobra.Command) (roleClient, error) { return nil, nil },
			func(cmd *cobra.Command) (resourceClient, error) { return nil, nil },
		)

		_, _, err := executeCommandForTest(t, cmd, "list", "--limit=-2")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--limit must be -1")
	})

	t.Run("--page-size flag overrides the server page size", func(t *testing.T) {
		listResources, calls := datasetListResources(5)
		mockClient := mockListClient{t: t, listResources: listResources}

		cmd := newListCmd(
			func(cmd *cobra.Command) (listClient, error) { return mockClient, nil },
			func(cmd *cobra.Command) (roleClient, error) { return nil, nil },
			func(cmd *cobra.Command) (resourceClient, error) { return nil, nil },
		)

		stdout, _, err := executeCommandForTest(t, cmd, "list", "--page-size=2", "--limit=-1", "-i")
		require.NoError(t, err)

		var got []map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(stdout), &got))
		assert.Len(t, got, 5)
		assert.GreaterOrEqual(t, *calls, 3, "expected multiple page requests (pages of 2 for 5 items)")
	})

	t.Run("rejects a non-positive --page-size", func(t *testing.T) {
		mockClient := mockListClient{t: t}
		cmd := newListCmd(
			func(cmd *cobra.Command) (listClient, error) { return mockClient, nil },
			func(cmd *cobra.Command) (roleClient, error) { return nil, nil },
			func(cmd *cobra.Command) (resourceClient, error) { return nil, nil },
		)

		_, _, err := executeCommandForTest(t, cmd, "list", "--page-size=0")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--page-size must be a positive integer")
	})

	t.Run("--page-size flag takes precedence over CONJUR_LIST_PAGE_SIZE", func(t *testing.T) {
		t.Setenv("CONJUR_LIST_PAGE_SIZE", "1")
		listResources, calls := datasetListResources(5)
		mockClient := mockListClient{t: t, listResources: listResources}

		cmd := newListCmd(
			func(cmd *cobra.Command) (listClient, error) { return mockClient, nil },
			func(cmd *cobra.Command) (roleClient, error) { return nil, nil },
			func(cmd *cobra.Command) (resourceClient, error) { return nil, nil },
		)

		stdout, _, err := executeCommandForTest(t, cmd, "list", "--page-size=5", "--limit=-1", "-i")
		require.NoError(t, err)

		var got []map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(stdout), &got))
		assert.Len(t, got, 5)
		assert.Equal(t, 2, *calls, "page-size=5 should fetch all 5 items in one full page plus one confirmation page, ignoring the env var's page size of 1")
	})

	t.Run("warns to stderr and falls back to default when CONJUR_LIST_PAGE_SIZE is invalid", func(t *testing.T) {
		t.Setenv("CONJUR_LIST_PAGE_SIZE", "not-a-number")
		listResources, _ := datasetListResources(5)
		mockClient := mockListClient{t: t, listResources: listResources}

		cmd := newListCmd(
			func(cmd *cobra.Command) (listClient, error) { return mockClient, nil },
			func(cmd *cobra.Command) (roleClient, error) { return nil, nil },
			func(cmd *cobra.Command) (resourceClient, error) { return nil, nil },
		)

		_, stderr, err := executeCommandForTest(t, cmd, "list", "-i")
		require.NoError(t, err)
		assert.Contains(t, stderr, "CONJUR_LIST_PAGE_SIZE")
	})

	t.Run("--page-size is hidden from help", func(t *testing.T) {
		cmd := newListCmd(
			func(cmd *cobra.Command) (listClient, error) { return nil, nil },
			func(cmd *cobra.Command) (roleClient, error) { return nil, nil },
			func(cmd *cobra.Command) (resourceClient, error) { return nil, nil },
		)

		flag := cmd.Flags().Lookup("page-size")
		require.NotNil(t, flag)
		assert.True(t, flag.Hidden, "--page-size is a rarely-needed tuning knob and shouldn't clutter --help")

		stdout, _, err := executeCommandForTest(t, cmd, "list", "--help")
		require.NoError(t, err)
		assert.NotContains(t, stdout, "--page-size")
	})

	t.Run("rejects --limit values below -1 on the --count path too, before contacting the server", func(t *testing.T) {
		calls := 0
		mockClient := mockListClient{
			t: t,
			listResources: func(t *testing.T, filter *conjurapi.ResourceFilter) ([]map[string]interface{}, error) {
				calls++
				return clientResponse, nil
			},
			resourcesCount: func(t *testing.T, filter *conjurapi.ResourceFilter) (*conjurapi.ResourcesCount, error) {
				calls++
				return &conjurapi.ResourcesCount{Count: 3}, nil
			},
		}
		cmd := newListCmd(
			func(cmd *cobra.Command) (listClient, error) { return mockClient, nil },
			func(cmd *cobra.Command) (roleClient, error) { return nil, nil },
			func(cmd *cobra.Command) (resourceClient, error) { return nil, nil },
		)

		_, _, err := executeCommandForTest(t, cmd, "list", "--count", "--limit=-2")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--limit must be -1")
		assert.Equal(t, 0, calls, "validation must reject before any API call, on the --count path too")
	})
}
