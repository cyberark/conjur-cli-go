package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockAuthenticatorClient struct {
	t                   *testing.T
	enableAuthenticator func(t *testing.T, authenticatorType string, serviceID string, enabled bool) error
}

func (m mockAuthenticatorClient) EnableAuthenticator(
	authenticatorType string,
	serviceID string,
	enabled bool,
) error {
	return m.enableAuthenticator(m.t, authenticatorType, serviceID, enabled)
}

type authenticatorCmdTestCase struct {
	name                string
	args                []string
	enableAuthenticator func(t *testing.T, authenticatorType string, serviceID string, enabled bool) error
	clientFactoryError  error
	assert              func(t *testing.T, stdout string, stderr string, err error)
}

var authenticatorCmdTestCases = []authenticatorCmdTestCase{
	{
		name: "authenticator command help",
		args: []string{"authenticator", "--help"},
		assert: func(t *testing.T, stdout string, stderr string, err error) {
			assert.Contains(t, stdout, "HELP LONG")
		},
	},
}

func sharedAuthenticatorCmdTestCases(
	subcommand string,
) []authenticatorCmdTestCase {
	return []authenticatorCmdTestCase{
		{
			name: fmt.Sprintf("%s subcommand help", subcommand),
			args: []string{"authenticator", subcommand, "--help"},
			assert: func(t *testing.T, stdout string, stderr string, err error) {
				assert.Contains(t, stdout, "HELP LONG")
			},
		},
		{
			name:               fmt.Sprintf("%s subcommand client factory error", subcommand),
			args:               []string{"authenticator", subcommand, "-i", "authn-iam/prod"},
			clientFactoryError: fmt.Errorf("client factory error"),
			assert: func(t *testing.T, stdout, stderr string, err error) {
				assert.Contains(t, stderr, "Error: client factory error")
			},
		},
		{
			name: fmt.Sprintf("%s subcommand missing id", subcommand),
			args: []string{"authenticator", subcommand},
			assert: func(t *testing.T, stdout, stderr string, err error) {
				assert.Contains(t, stderr, "Error: required flag(s) \"id\" not set\n")
			},
		},
		{
			name: fmt.Sprintf("%s subcommand invalid authenticator type", subcommand),
			args: []string{"authenticator", subcommand, "-i", "invalid-authenticator-type"},
			assert: func(t *testing.T, stdout, stderr string, err error) {
				assert.Contains(t, stderr, "Error: invalid authenticator ID format: invalid-authenticator-type, expected format is 'authenticator_type/service_id'\n")
			},
		},
		{
			name: fmt.Sprintf("%s subcommand supports authn-gcp as id", subcommand),
			args: []string{"authenticator", subcommand, "-i", "authn-gcp"},
			enableAuthenticator: func(t *testing.T, authenticatorType string, serviceID string, enabled bool) error {
				return nil
			},
			assert: func(t *testing.T, stdout, stderr string, err error) {
				assert.NotContains(t, stderr, "Error: invalid authenticator ID format: authn-gcp, expected format is 'authenticator_type/service_id'\n")
			},
		},
		{
			name: fmt.Sprintf("%s subcommand with good response", subcommand),
			args: []string{"authenticator", subcommand, "-i", "authn-iam/prod"},
			enableAuthenticator: func(t *testing.T, authenticatorType string, serviceID string, enabled bool) error {
				return nil
			},
			assert: func(t *testing.T, stdout, stderr string, err error) {
				msg := fmt.Sprintf("Authenticator authn-iam/prod is %sd\n", subcommand)
				assert.Equal(t, msg, stdout)
			},
		},
		{
			name: fmt.Sprintf("%s subcommand response error", subcommand),
			args: []string{"authenticator", subcommand, "-i", "authn-iam/prod"},
			enableAuthenticator: func(t *testing.T, authenticatorType string, serviceID string, enabled bool) error {
				return fmt.Errorf("some error")
			},
			assert: func(t *testing.T, stdout, stderr string, err error) {
				assert.Contains(t, stderr, "Error: some error")
			},
		},
	}
}

func TestAuthenticatorCmd(t *testing.T) {
	t.Parallel()

	var allTests []authenticatorCmdTestCase
	for _, cases := range [][]authenticatorCmdTestCase{
		authenticatorCmdTestCases,
		sharedAuthenticatorCmdTestCases("enable"),
		sharedAuthenticatorCmdTestCases("disable"),
	} {
		allTests = append(allTests, cases...)
	}

	for _, tc := range allTests {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := mockAuthenticatorClient{
				t:                   t,
				enableAuthenticator: tc.enableAuthenticator,
			}
			cmd := newAuthenticatorCommand(
				func(cmd *cobra.Command) (authenticatorClient, error) {
					return mockClient, tc.clientFactoryError
				})

			stdout, stderr, err := executeCommandForTest(
				t, cmd, tc.args...,
			)
			if tc.assert != nil {
				tc.assert(t, stdout, stderr, err)
			}

		})
	}
}
