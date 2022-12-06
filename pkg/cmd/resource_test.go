package cmd

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type mockResourceClient struct {
	t        *testing.T
	resource func(t *testing.T, resourceID string) (resource map[string]interface{}, err error)
}

func (m mockResourceClient) Resource(resourceID string) (resource map[string]interface{}, err error) {
	return m.resource(m.t, resourceID)
}

var resourceExistsCmdTestCases = []struct {
	name               string
	args               []string
	resource           func(t *testing.T, resourceID string) (resource map[string]interface{}, err error)
	clientFactoryError error
	assert             func(t *testing.T, stdout string, stderr string, err error)
}{
	{
		name: "resource exists help",
		args: []string{"exists", "--help"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "Checks if whether a resource exists")
		},
	},
	{
		name: "resource exists missing resource-id",
		args: []string{"exists"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "This command requires one argument, a [resource-id].")
		},
	},
	{
		name: "resource exists return true",
		args: []string{"exists", "meow"},
		resource: func(t *testing.T, resourceID string) (resource map[string]interface{}, err error) {
			assert.Equal(t, "meow", resourceID)

			return map[string]interface{}{"key": "value"}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "true")
		},
	},
	{
		name: "resource exists return false",
		args: []string{"exists", "meow"},
		resource: func(t *testing.T, resourceID string) (resource map[string]interface{}, err error) {
			assert.Equal(t, "meow", resourceID)

			return nil, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "false")
		},
	},
	{
		name: "resource exists client error",
		args: []string{"exists", "abcdefg"},
		resource: func(t *testing.T, resourceID string) (resource map[string]interface{}, err error) {
			return nil, fmt.Errorf("%s", "an error")
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

func TestResourceExistsCmd(t *testing.T) {
	for _, tc := range resourceExistsCmdTestCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := mockResourceClient{t: t, resource: tc.resource}

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
