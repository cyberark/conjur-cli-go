package cmd

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type mockAuthenticateClient struct {
	internalAuthenticate func() ([]byte, error)
}

func (m mockAuthenticateClient) InternalAuthenticate() ([]byte, error) {
	return m.internalAuthenticate()
}

var authenticateCmdTestCases = []struct {
	name                 string
	args                 []string
	internalAuthenticate func() ([]byte, error)
	clientFactoryError   error
	assert               func(t *testing.T, stdout, stderr string, err error)
}{
	{
		name: "display help",
		args: []string{"authenticate", "--help"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "json response",
		args: []string{"authenticate"},
		internalAuthenticate: func() ([]byte, error) {
			return []byte(`{"user":"test"}`), nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			expectedOut := `{
  "user": "test"
}
`

			assert.Contains(t, stdout, expectedOut)
		},
	},
	{
		name: "non-json response",
		args: []string{"authenticate"},
		internalAuthenticate: func() ([]byte, error) {
			return []byte(`not json response`), nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: invalid")
		},
	},
	{
		name: "header format: json response",
		args: []string{"authenticate", "-H"},
		internalAuthenticate: func() ([]byte, error) {
			return []byte(`{"user":"test"}`), nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			expectedOut := `Authorization: Token token="eyJ1c2VyIjoidGVzdCJ9"
`

			assert.Contains(t, stdout, expectedOut)
		},
	},
	{
		name: "header format: non json response",
		args: []string{"authenticate", "-H"},
		internalAuthenticate: func() ([]byte, error) {
			return []byte(`non json response`), nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			expectedOut := `Authorization: Token token="bm9uIGpzb24gcmVzcG9uc2U="
`

			assert.Contains(t, stdout, expectedOut)
		},
	},
	{
		name: "client error",
		args: []string{"authenticate"},
		internalAuthenticate: func() ([]byte, error) {
			return nil, fmt.Errorf("%s", "an error")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: an error\n")
		},
	},
	{
		name:               "client factory error",
		args:               []string{"authenticate"},
		clientFactoryError: fmt.Errorf("%s", "client factory error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: client factory error\n")
		},
	},
}

func TestAuthenticateCmd(t *testing.T) {
	t.Parallel()

	for _, tc := range authenticateCmdTestCases {
		t.Run(tc.name, func(t *testing.T) {
			testWhoamiClientFactory := func(cmd *cobra.Command) (authenticateClient, error) {
				return mockAuthenticateClient{internalAuthenticate: tc.internalAuthenticate}, tc.clientFactoryError
			}
			cmd := newAuthenticateCommand(testWhoamiClientFactory)

			stdout, stderr, err := executeCommandForTest(t, cmd, tc.args...)
			tc.assert(t, stdout, stderr, err)
		})
	}
}
