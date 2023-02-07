package clients

import (
	"net/http"
	"testing"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-api-go/conjurapi/authn"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestMaybeDebugLoggingForClient(t *testing.T) {
	var debugTestCases = []struct {
		name   string
		debug  bool
		assert func(t *testing.T, client *conjurapi.Client)
	}{
		{
			name:  "debug is true enables HTTP logging",
			debug: true,
			assert: func(t *testing.T, client *conjurapi.Client) {
				assert.NotNil(t, client.GetHttpClient().Transport)
			},
		},
		{
			name:  "debug is false uses the default transport",
			debug: false,
			assert: func(t *testing.T, client *conjurapi.Client) {
				assert.Nil(t, client.GetHttpClient().Transport)
			},
		},
	}

	for _, tc := range debugTestCases {
		t.Run(tc.name, func(t *testing.T) {
			client, _ := conjurapi.NewClientFromKey(conjurapi.Config{Account: "conjur", ApplianceURL: "http://conjur.com"}, authn.LoginPair{"username", "password"})
			client.SetHttpClient(&http.Client{})
			cmd := &cobra.Command{}
			MaybeDebugLoggingForClient(tc.debug, cmd, client)

			tc.assert(t, client)
		})
	}
}
