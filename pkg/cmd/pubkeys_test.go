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
