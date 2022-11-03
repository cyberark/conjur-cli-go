package cmd

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type mockCheckClient struct {
	t               *testing.T
	checkPermission func(t *testing.T, resourceID, privilege string) (bool, error)
}

func (m mockCheckClient) CheckPermission(resourceID, privilege string) (bool, error) {
	return m.checkPermission(m.t, resourceID, privilege)
}

var checkCmdTestCases = []struct {
	name               string
	args               []string
	checkPermission    func(t *testing.T, resourceID, privilege string) (bool, error)
	clientFactoryError error
	assert             func(t *testing.T, stdout string, stderr string, err error)
}{
	{
		name: "check help",
		args: []string{"check", "--help"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "This command requires two arguments, a [resourceId] and a [privilege].")
		},
	},
	{
		name: "check no args",
		args: []string{"check"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "This command requires two arguments, a [resourceId] and a [privilege].")
		},
	},
	{
		name: "check missing resourceID",
		args: []string{"check", "write"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "This command requires two arguments, a [resourceId] and a [privilege].")
		},
	},
	{
		name: "check missing privilege",
		args: []string{"check", "dev:variable:somevariable"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "This command requires two arguments, a [resourceId] and a [privilege].")
		},
	},
	{
		name: "check return true",
		args: []string{"check", "dev:variable:secret", "read"},
		checkPermission: func(t *testing.T, resourceID, privilege string) (bool, error) {
			assert.Equal(t, "dev:variable:secret", resourceID)
			assert.Equal(t, "read", privilege)

			return true, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "true")
		},
	},
	{
		name: "check return false",
		args: []string{"check", "dev:variable:secret", "write"},
		checkPermission: func(t *testing.T, resourceID, privilege string) (bool, error) {
			assert.Equal(t, "dev:variable:secret", resourceID)
			assert.Equal(t, "write", privilege)

			return false, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "false")
		},
	},
	{
		name: "check client error",
		args: []string{"check", "abcdefg", "hijklmn"},
		checkPermission: func(t *testing.T, resourceID, privilege string) (bool, error) {
			return false, fmt.Errorf("%s", "an error")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: an error\n")
		},
	},
	{
		name:               "check client factory error",
		args:               []string{"check", "abcdefg", "hijklmn"},
		clientFactoryError: fmt.Errorf("%s", "client factory error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: client factory error\n")
		},
	},
}

func TestCheckCmd(t *testing.T) {
	for _, tc := range checkCmdTestCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := mockCheckClient{t: t, checkPermission: tc.checkPermission}

			cmd := newCheckCmd(
				func(cmd *cobra.Command) (checkClient, error) {
					return mockClient, tc.clientFactoryError
				},
			)

			stdout, stderr, err := executeCommandForTest(t, cmd, tc.args...)
			tc.assert(t, stdout, stderr, err)
		})
	}
}
