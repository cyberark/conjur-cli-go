package cmd

import (
	"fmt"
	"testing"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type mockListClient struct {
	t             *testing.T
	listResources func(t *testing.T, filter *conjurapi.ResourceFilter) ([]string, error)
}

func (m mockListClient) ResourceIDs(filter *conjurapi.ResourceFilter) ([]string, error) {
	return m.listResources(m.t, filter)
}

var listCmdTestCases = []struct {
	name               string
	args               []string
	listResources      func(t *testing.T, filter *conjurapi.ResourceFilter) ([]string, error)
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
		listResources: func(t *testing.T, filter *conjurapi.ResourceFilter) ([]string, error) {
			assert.Equal(t, "woof", filter.Kind)

			return []string{"abc"}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "abc")
		},
	},
	{
		name: "list search",
		args: []string{"list", "-s", "friend"},
		listResources: func(t *testing.T, filter *conjurapi.ResourceFilter) ([]string, error) {
			assert.Equal(t, "friend", filter.Search)

			return []string{"abc"}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "abc")
		},
	},
	{
		name: "list limit",
		args: []string{"list", "-l", "5"},
		listResources: func(t *testing.T, filter *conjurapi.ResourceFilter) ([]string, error) {
			assert.Equal(t, 5, filter.Limit)

			return []string{"abc"}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "abc")
		},
	},
	{
		name: "list offset",
		args: []string{"list", "-o", "2"},
		listResources: func(t *testing.T, filter *conjurapi.ResourceFilter) ([]string, error) {
			assert.Equal(t, 2, filter.Offset)

			return []string{"abc"}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "abc")
		},
	},
	{
		name: "list no flags",
		args: []string{"list"},
		listResources: func(t *testing.T, filter *conjurapi.ResourceFilter) ([]string, error) {
			assert.Equal(t, "", filter.Kind)
			assert.Equal(t, "", filter.Search)
			assert.Equal(t, 0, filter.Limit)
			assert.Equal(t, 0, filter.Offset)

			return []string{"abc"}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "abc")
		},
	},
	{
		name: "list all flags",
		args: []string{"list", "-k", "qwerty", "-s", "abc", "-l", "5", "-o", "2"},
		listResources: func(t *testing.T, filter *conjurapi.ResourceFilter) ([]string, error) {
			assert.Equal(t, "qwerty", filter.Kind)
			assert.Equal(t, "abc", filter.Search)
			assert.Equal(t, 5, filter.Limit)
			assert.Equal(t, 2, filter.Offset)

			return []string{"abc"}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "abc")
		},
	},
	{
		name: "list pretty format",
		args: []string{"list"},
		listResources: func(t *testing.T, filter *conjurapi.ResourceFilter) ([]string, error) {
			return []string{"moo", "meow"}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "[\n  \"moo\",\n  \"meow\"\n]\n")
		},
	},
	{
		name: "list client error",
		args: []string{"list", "-k", "asdf", "-s", "qwer"},
		listResources: func(t *testing.T, filter *conjurapi.ResourceFilter) ([]string, error) {
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
}

func TestListCmd(t *testing.T) {
	for _, tc := range listCmdTestCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := mockListClient{t: t, listResources: tc.listResources}

			cmd := newListCmd(
				func(cmd *cobra.Command) (listClient, error) {
					return mockClient, tc.clientFactoryError
				},
			)

			stdout, stderr, err := executeCommandForTest(t, cmd, tc.args...)
			tc.assert(t, stdout, stderr, err)
		})
	}
}
