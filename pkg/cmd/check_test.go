package cmd

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type mockCheckClient struct {
	t                      *testing.T
	checkPermission        func(t *testing.T, resourceID string, privilege string) (bool, error)
	checkPermissionForRole func(t *testing.T, resourceID string, roleID string, privilege string) (bool, error)
}

func (m mockCheckClient) CheckPermission(resourceID string, privilege string) (bool, error) {
	return m.checkPermission(m.t, resourceID, privilege)
}

func (m mockCheckClient) CheckPermissionForRole(resourceID string, roleID string, privilege string) (bool, error) {
	return m.checkPermissionForRole(m.t, resourceID, roleID, privilege)
}

var checkCmdTestCases = []struct {
	name                   string
	args                   []string
	checkPermission        func(t *testing.T, resourceID string, privilege string) (bool, error)
	checkPermissionForRole func(t *testing.T, resourceID string, roleID string, privilege string) (bool, error)
	clientFactoryError     error
	assert                 func(t *testing.T, stdout string, stderr string, err error)
}{
	{
		name: "check help",
		args: []string{"check", "--help"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "check no args",
		args: []string{"check"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "check missing resourceID",
		args: []string{"check", "write"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "check missing privilege",
		args: []string{"check", "dev:variable:somevariable"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "check with role flag and missing params",
		args: []string{"check", "-r", "dev:user:alice"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "check returns true for default admin role",
		args: []string{"check", "dev:variable:secret", "read"},
		checkPermission: func(t *testing.T, resourceID string, privilege string) (bool, error) {
			assert.Equal(t, "dev:variable:secret", resourceID)
			assert.Equal(t, "read", privilege)

			return true, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "true")
		},
	},
	{
		name: "check returns false for default admin role",
		args: []string{"check", "dev:variable:secret", "write"},
		checkPermission: func(t *testing.T, resourceID string, privilege string) (bool, error) {
			assert.Equal(t, "dev:variable:secret", resourceID)
			assert.Equal(t, "write", privilege)

			return false, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "false")
		},
	},
	{
		name: "check returns true for specific role",
		args: []string{"check", "-r", "dev:user:alice", "dev:variable:secret", "read"},
		checkPermissionForRole: func(t *testing.T, resourceID string, roleID string, privilege string) (bool, error) {
			assert.Equal(t, "dev:variable:secret", resourceID)
			assert.Equal(t, "dev:user:alice", roleID)
			assert.Equal(t, "read", privilege)

			return true, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "true")
		},
	},
	{
		name: "check returns false for specific role",
		args: []string{"check", "-r", "dev:user:alice", "dev:variable:secret", "write"},
		checkPermissionForRole: func(t *testing.T, resourceID string, roleID string, privilege string) (bool, error) {
			assert.Equal(t, "dev:variable:secret", resourceID)
			assert.Equal(t, "dev:user:alice", roleID)
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
		checkPermission: func(t *testing.T, resourceID string, privilege string) (bool, error) {
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
			mockClient := mockCheckClient{
				t:                      t,
				checkPermission:        tc.checkPermission,
				checkPermissionForRole: tc.checkPermissionForRole,
			}

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
