package cmd

import (
	"fmt"
	"github.com/cyberark/conjur-api-go/conjurapi"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type mockHFClient struct {
	t      *testing.T
	create func(*testing.T, string, string, []string, int) ([]conjurapi.HostFactoryTokenResponse, error)
	revoke func(*testing.T, string) error
	host   func(*testing.T, string, string) (conjurapi.HostFactoryHostResponse, error)
}

func (m mockHFClient) CreateToken(durationStr string, hostFactory string, cidrs []string, count int) ([]conjurapi.HostFactoryTokenResponse, error) {
	return m.create(m.t, durationStr, hostFactory, cidrs, count)
}
func (m mockHFClient) DeleteToken(token string) error {
	return m.revoke(m.t, token)
}

func (m mockHFClient) CreateHost(id string, token string) (conjurapi.HostFactoryHostResponse, error) {
	return m.host(m.t, id, token)
}

var hostfactoryCmdTestCases = []struct {
	name               string
	args               []string
	create             func(t *testing.T, duration string, hostFactory string, cidrs []string, count int) ([]conjurapi.HostFactoryTokenResponse, error)
	revoke             func(t *testing.T, token string) error
	host               func(t *testing.T, id string, token string) (conjurapi.HostFactoryHostResponse, error)
	clientFactoryError error
	assert             func(t *testing.T, stdout string, stderr string, err error)
}{
	{
		name: "hostfactory command help",
		args: []string{"hostfactory", "--help"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "token command help",
		args: []string{"hostfactory", "tokens", "--help"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "token create command help",
		args: []string{"hostfactory", "tokens", "create", "--help"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "token revoke command help",
		args: []string{"hostfactory", "tokens", "revoke", "--help"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "host command help",
		args: []string{"hostfactory", "host", "--help"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "host create command help",
		args: []string{"hostfactory", "host", "create", "--help"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "token create command success",
		args: []string{"hostfactory", "tokens", "create", "--duration", "5m", "-f", "cucumber_host_factory_factory"},
		create: func(t *testing.T, duration string, hostFactory string, cidrs []string, count int) ([]conjurapi.HostFactoryTokenResponse, error) {
			return []conjurapi.HostFactoryTokenResponse{
				conjurapi.HostFactoryTokenResponse{
					Expiration: "2022-12-23T20:32:46Z",
					Cidr:       []string{"0.0.0.0/32"},
					Token:      "1bfpyr3y41kb039ykpyf2hm87ez2dv9hdc3r5sh1n2h9z7j22mga2da"},
			}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "1bfpyr3y41kb039ykpyf2hm87ez2dv9hdc3r5sh1n2h9z7j22mga2da")
		},
	},
	{
		name: "token create with ip success",
		args: []string{"hostfactory", "tokens", "create", "--duration", "5m", "-f", "cucumber_host_factory_factory",
			"-c", "0.0.0.0,1.2.3.4"},
		create: func(t *testing.T, duration string, hostFactory string, cidrs []string, count int) ([]conjurapi.HostFactoryTokenResponse, error) {
			return []conjurapi.HostFactoryTokenResponse{
				conjurapi.HostFactoryTokenResponse{
					Expiration: "2022-12-23T20:32:46Z",
					Cidr:       []string{"0.0.0.0/32", "1.2.3.4/32"},
					Token:      "1bfpyr3y41kb039ykpyf2hm87ez2dv9hdc3r5sh1n2h9z7j22mga2da"},
			}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "1bfpyr3y41kb039ykpyf2hm87ez2dv9hdc3r5sh1n2h9z7j22mga2da")
			assert.Contains(t, stdout, "[\n      \"0.0.0.0/32\",\n      \"1.2.3.4/32\"\n    ]")
		},
	},
	{
		name: "token create command error",
		args: []string{"hostfactory", "tokens", "create", "-d", "5m", "-f", "cucumber_host_factory_factory",
			"-c", "0.0.0"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "invalid string being converted")
		},
	},
	{
		name: "token create missing flag",
		args: []string{"hostfactory", "tokens", "create"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: required flag(s) \"hostFactory\" not set")
		},
	},
	{
		name: "token revoke command success",
		args: []string{"hostfactory", "tokens", "revoke", "-t", "1bfpyr3y41kb039ykpyf2hm87ez2dv9hdc3r5sh1n2h9z7j22mga2da"},
		revoke: func(t *testing.T, token string) error {
			return nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "")
		},
	},
	{
		name: "token revoke command error",
		args: []string{"hostfactory", "tokens", "revoke", "-t", "12345"},
		revoke: func(t *testing.T, token string) error {
			return fmt.Errorf("%s", "404 Not Found. Host_factory_token '12345' not found in account")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: 404 Not Found. Host_factory_token '12345' not found in account")
		},
	},
	{
		name: "token revoke missing flag",
		args: []string{"hostfactory", "tokens", "revoke", "1bfpyr3y41kb039ykpyf2hm87ez2dv9hdc3r5sh1n2h9z7j22mga2da"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: required flag(s) \"token\" not set")
		},
	},
	{
		name: "host create success",
		args: []string{"hostfactory", "hosts", "create", "-id", "new-host", "-t", "1bfpyr3y41kb039ykpyf2hm87ez2dv9hdc3r5sh1n2h9z7j22mga2da"},
		host: func(t *testing.T, id string, token string) (conjurapi.HostFactoryHostResponse, error) {
			return conjurapi.HostFactoryHostResponse{
				CreatedAt: "2023-01-01",
				Id:        "new-host",
				ApiKey:    "1234567890",
			}, nil
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "1234567890")
		},
	},
	{
		name: "host create error",
		args: []string{"hostfactory", "hosts", "create", "-id", "new-host", "-t", "notvalidtoken"},
		host: func(t *testing.T, id string, token string) (conjurapi.HostFactoryHostResponse, error) {
			return conjurapi.HostFactoryHostResponse{}, fmt.Errorf("%s", "401 Unauthorized.")
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: 401 Unauthorized.")
		},
	},
	{
		name: "host create missing flag",
		args: []string{"hostfactory", "hosts", "create", "-id", "new-host"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Error: required flag(s) \"token\" not set")
		},
	},
}

func TestHostfactoryCmd(t *testing.T) {
	t.Parallel()
	for _, tc := range hostfactoryCmdTestCases {
		t.Run(tc.name, func(t *testing.T) {
			testCreateTokenClient := func(cmd *cobra.Command) (createTokenClient, error) {
				return mockHFClient{t: t, create: tc.create}, tc.clientFactoryError
			}
			testRevokeTokenClient := func(cmd *cobra.Command) (revokeTokenClient, error) {
				return mockHFClient{t: t, revoke: tc.revoke}, tc.clientFactoryError
			}
			testCreateHostClient := func(cmd *cobra.Command) (createHostClient, error) {
				return mockHFClient{t: t, host: tc.host}, tc.clientFactoryError
			}
			cmd := newHostFactoryCmd(testCreateTokenClient, testRevokeTokenClient, testCreateHostClient)
			stdout, stderr, err := executeCommandForTest(t, cmd, tc.args...)
			tc.assert(t, stdout, stderr, err)
		})
	}
}