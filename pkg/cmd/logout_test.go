package cmd

import (
	"testing"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/stretchr/testify/assert"
)

var logoutTestCases = []struct {
	name   string
	args   []string
	config conjurapi.Config
	assert func(t *testing.T, stdout string, stderr string, err error)
}{
	{
		name:   "logout command help",
		args:   []string{"logout", "--help"},
		config: conjurapi.Config{},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name:   "logout",
		args:   []string{"logout"},
		config: conjurapi.Config{},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.NoError(t, err)
			assert.Contains(t, stdout, "Logged out")
			assert.Empty(t, stderr)
		},
	},
}

func TestLogoutCmd(t *testing.T) {
	for _, tc := range logoutTestCases {
		// Mock out the configuration loader
		configLoader := func() (conjurapi.Config, error) {
			return tc.config, nil
		}

		t.Run(tc.name, func(t *testing.T) {
			cmd := newLogoutCmd(configLoader)

			stdout, stderr, err := executeCommandForTest(t, cmd, tc.args...)
			tc.assert(t, stdout, stderr, err)
		})
	}
}
