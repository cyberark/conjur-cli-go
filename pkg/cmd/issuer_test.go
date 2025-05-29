package cmd

import (
	"errors"
	"testing"

	api "github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type mockIssuerClient struct {
	createIssuer func(api.Issuer) (api.Issuer, error)
}

func (m *mockIssuerClient) CreateIssuer(issuer api.Issuer) (api.Issuer, error) {
	return m.createIssuer(issuer)
}

var createCmdTestCases = []struct {
	name               string
	args               []string
	access_key         string
	secret_key         string
	issuerFactoryError error
	createIssuerError  error
	assert             func(t *testing.T, stdout string, stderr string, err error)
}{
	{
		name: "without subcommand",
		args: []string{},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "with help flag",
		args: []string{"--help"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
	{
		name: "missing --data flag",
		args: []string{"create", "--id", "test-issuer", "--type", "aws", "--max-ttl", "3000"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "required flag(s) \"data\" not set")
		},
	},
	{
		name: "missing --id flag",
		args: []string{"create", "--type", "aws", "--max-ttl", "3000", "--data", `{"key":"value"}`},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "required flag(s) \"id\" not set")
		},
	},
	{
		name: "missing --max-ttl flag",
		args: []string{"create", "--id", "test-issuer", "--type", "aws", "--data", `{"key":"value"}`},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "required flag(s) \"max-ttl\" not set")
		},
	},
	{
		name: "missing --type flag",
		args: []string{"create", "--id", "test-issuer", "--max-ttl", "3000", "--data", `{"key":"value"}`},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "required flag(s) \"type\" not set")
		},
	},
	{
		name: "invalid --max-ttl type",
		args: []string{"create", "--id", "test-issuer", "--type", "aws", "--max-ttl", "invalid", "--data", `{"key":"value"}`},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "invalid argument \"invalid\" for \"-m, --max-ttl\" flag")
		},
	},
	{
		name: "invalid --data type",
		args: []string{"create", "--id", "test-issuer", "--type", "aws", "--max-ttl", "3000", "--data", "invalid-json"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Failed to parse 'data' flag JSON:")
		},
	},
	{
		name:               "clientFactory returns an error",
		args:               []string{"create", "--id", "test-issuer", "--type", "aws", "--max-ttl", "3000", "--data", `{"key":"value"}`},
		issuerFactoryError: errors.New("mock issuerFactory error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "mock issuerFactory error")
		},
	},
	{
		name:              "CreateIssuer function returns an error",
		args:              []string{"create", "--id", "test-issuer", "--type", "aws", "--max-ttl", "3000", "--data", `{"key":"value"}`},
		createIssuerError: errors.New("mock CreateIssuer error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "mock CreateIssuer error")
		},
	},
	{
		name: "issuer is created",
		args: []string{"create", "--id", "test-issuer", "--type", "aws", "--max-ttl", "3000", "--data", `{"key":"value"}`},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			output := "{\n  \"id\": \"test-issuer\",\n  \"type\": \"aws\",\n  \"max_ttl\": 3000,\n  \"data\": {\n    \"key\": \"value\"\n  }\n}\n"
			assert.Contains(t, stdout, output)
		},
	},
}

func TestCreateIssuerCmd(t *testing.T) {
	for _, tc := range createCmdTestCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := newCreateIssuerCmd(
				func(cmd *cobra.Command) (issuerClient, error) {
					return &mockIssuerClient{
						createIssuer: func(issuer api.Issuer) (api.Issuer, error) {
							return issuer, tc.createIssuerError
						},
					}, tc.issuerFactoryError
				},
			)

			stdout, stderr, err := executeCommandForTest(t, cmd, tc.args...)
			tc.assert(t, stdout, stderr, err)
		})
	}
}
