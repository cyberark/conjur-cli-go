package cmd

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type mockPubKeysClient struct {
	t       *testing.T
	pubKeys func(t *testing.T, kind string, identifier string) ([]byte, error)
}

func (m mockPubKeysClient) PublicKeys(kind string, identifier string) ([]byte, error) {
	return m.pubKeys(m.t, kind, identifier)
}

var pubKeysCmdTestCases = []struct {
	name               string
	args               []string
	pubKeys            func(t *testing.T, kind string, identifier string) ([]byte, error)
	clientFactoryError error
	assert             func(t *testing.T, stdout, stderr string, err error)
}{
	{
		name: "display help",
		args: []string{"pubkeys", "--help"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "display user public keys",
		args: []string{"pubkeys", "alice"},
		pubKeys: func(t *testing.T, kind string, identifier string) ([]byte, error) {
			return []byte(`
ssh-rsa test-key laptop

`), nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			expectedOut := `
ssh-rsa test-key laptop

`
			assert.Contains(t, stdout, expectedOut)
		},
	},
	{
		name: "client error",
		args: []string{"pubkeys", "alice"},
		pubKeys: func(t *testing.T, kind string, identifier string) ([]byte, error) {
			return nil, fmt.Errorf("%s", "an error")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: an error\n")
		},
	},
	{
		name: "missing pubkeys endpoint returns friendly error for 404",
		args: []string{"pubkeys", "alice"},
		pubKeys: func(t *testing.T, kind string, identifier string) ([]byte, error) {
			return nil, fmt.Errorf("404 Not Found")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Empty(t, stdout)
			assert.Contains(t, stderr, "Error: public keys endpoint is not available on this server: the server may not support this feature\n")
		},
	},
	{
		name: "missing pubkeys endpoint returns friendly error for rails route mismatch",
		args: []string{"pubkeys", "alice"},
		pubKeys: func(t *testing.T, kind string, identifier string) ([]byte, error) {
			return nil, fmt.Errorf("No route matches [GET] '/authn/account/user/alice/public_keys'")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Empty(t, stdout)
			assert.Contains(t, stderr, "Error: public keys endpoint is not available on this server: the server may not support this feature\n")
		},
	},
	{
		name: "missing pubkeys endpoint returns friendly error when conjur-api-go already translated the 404",
		args: []string{"pubkeys", "alice"},
		pubKeys: func(t *testing.T, kind string, identifier string) ([]byte, error) {
			// conjur-api-go translates the raw 404 into this message before the CLI sees it
			return nil, fmt.Errorf("public keys endpoint is not available on this server (got 404): the server may not support this feature")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Empty(t, stdout)
			assert.Contains(t, stderr, "Error: public keys endpoint is not available on this server: the server may not support this feature\n")
		},
	},
	{
		name:               "client factory error",
		args:               []string{"pubkeys", "alice"},
		clientFactoryError: fmt.Errorf("%s", "client factory error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: client factory error\n")
		},
	},
}

func TestPubKeysCmd(t *testing.T) {
	t.Parallel()

	for _, tc := range pubKeysCmdTestCases {
		t.Run(tc.name, func(t *testing.T) {
			testPubKeysClientFactory := func(cmd *cobra.Command) (pubKeysClient, error) {
				return mockPubKeysClient{t: t, pubKeys: tc.pubKeys}, tc.clientFactoryError
			}
			cmd := newPubKeysCommand(testPubKeysClientFactory)
			stdout, stderr, err := executeCommandForTest(t, cmd, tc.args...)
			tc.assert(t, stdout, stderr, err)
		})
	}
}
