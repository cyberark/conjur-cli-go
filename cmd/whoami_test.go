package cmd

import (
	"testing"

	"github.com/cyberark/conjur-cli-go/pkg/test"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type testWhoamiClient struct{}

func (tc testWhoamiClient) WhoAmI() ([]byte, error) {
	return []byte("{\"user\":\"test\"}"), nil
}

func testWhoamiClientFactory(cmd *cobra.Command) (whoamiClient, error) {
	return testWhoamiClient{}, nil
}

func TestWhoamiCmd(t *testing.T) {
	tt := []struct {
		args []string
		out  string
	}{
		{
			args: []string{"--help"},
			out:  "Displays info about the logged in user",
		},
		{
			args: []string{"--help=false"},
			out: "{\n" +
				"  \"user\": \"test\"\n" +
				"}",
		},
	}

	cmd := NewWhoamiCommand(testWhoamiClientFactory)

	for _, tc := range tt {
		out, _ := test.Execute(t, cmd, tc.args...)
		assert.Contains(t, out, tc.out)
	}
}
