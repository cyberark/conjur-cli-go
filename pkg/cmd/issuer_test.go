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
	updateIssuer func(issuerID string, issuerUpdate api.IssuerUpdate) (updated api.Issuer, err error)
	deleteIssuer func(issuerID string, keepSecrets bool) (err error)
	issuer       func(issuerID string) (issuer api.Issuer, err error)
	issuers      func() (issuer []api.Issuer, err error)
}

func (m *mockIssuerClient) CreateIssuer(issuer api.Issuer) (api.Issuer, error) {
	return m.createIssuer(issuer)
}

func (m *mockIssuerClient) DeleteIssuer(issuerID string, keepSecrets bool) (err error) {
	return m.deleteIssuer(issuerID, keepSecrets)
}
func (m *mockIssuerClient) Issuers() (issuers []api.Issuer, err error) {
	return m.issuers()
}
func (m *mockIssuerClient) Issuer(issuerID string) (issuer api.Issuer, err error) {
	return m.issuer(issuerID)
}

func (m *mockIssuerClient) UpdateIssuer(issuerID string, issuerUpdate api.IssuerUpdate) (updated api.Issuer, err error) {
	return m.updateIssuer(issuerID, issuerUpdate)
}

var createCmdTestCases = []struct {
	name               string
	getStringErr       map[string]interface{}
	getIntErr          error
	args               []string
	clientFactoryError error
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
		clientFactoryError: errors.New("mock issuerFactory error"),
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
		name:         "getStringFlagFunc retunrs an error for id key",
		args:         []string{"create", "--id", "test-issuer", "--type", "aws", "--max-ttl", "3000", "--data", `{"key":"value"}`},
		getStringErr: map[string]interface{}{"error": errors.New("mock getString error"), "key": "id"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "mock getString error")
		},
	},
	{
		name:         "getStringFlagFunc an error for type key",
		args:         []string{"create", "--id", "test-issuer", "--type", "aws", "--max-ttl", "3000", "--data", `{"key":"value"}`},
		getStringErr: map[string]interface{}{"error": errors.New("mock getString error"), "key": "type"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "mock getString error")
		},
	},
	{
		name:         "getStringFlagFunc returns an error for data key",
		args:         []string{"create", "--id", "test-issuer", "--type", "aws", "--max-ttl", "3000", "--data", `{"key":"value"}`},
		getStringErr: map[string]interface{}{"error": errors.New("mock getString error"), "key": "data"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "mock getString error")
		},
	},
	{
		name:      "getIntFlagFunc an error",
		args:      []string{"create", "--id", "test-issuer", "--type", "aws", "--max-ttl", "3000", "--data", `{"key":"value"}`},
		getIntErr: errors.New("mock getInt error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "mock getInt error")
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
		// Mock dependencies
		oldIntFunc := getIntFlagFunc
		oldStringFunc := getStringFlagFunc

		getIntFlagFunc = func(cmd *cobra.Command, key string) (int, error) {
			if tc.getIntErr != nil {
				return 0, tc.getIntErr
			}

			return cmd.Flags().GetInt(key)
		}
		getStringFlagFunc = func(cmd *cobra.Command, key string) (string, error) {
			if tc.getStringErr != nil && key == tc.getStringErr["key"] {
				return "", tc.getStringErr["error"].(error)
			}

			return cmd.Flags().GetString(key)
		}
		defer func() {
			getIntFlagFunc = oldIntFunc
			getStringFlagFunc = oldStringFunc
		}()

		t.Run(tc.name, func(t *testing.T) {
			cmd := newCreateIssuerCmd(
				func(cmd *cobra.Command) (issuerClient, error) {
					return &mockIssuerClient{
						createIssuer: func(issuer api.Issuer) (api.Issuer, error) {
							return issuer, tc.createIssuerError
						},
					}, tc.clientFactoryError
				},
			)

			stdout, stderr, err := executeCommandForTest(t, cmd, tc.args...)
			tc.assert(t, stdout, stderr, err)
		})
	}
}

var updateCmdTestCases = []struct {
	name               string
	getStringErr       map[string]interface{}
	getIntErr          error
	args               []string
	clientFactoryError error
	updateIssuerError  error
	issuer             api.Issuer
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
		name: "missing --id flag",
		args: []string{"update", "--max-ttl", "3000", "--data", `{"key":"value"}`},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "required flag(s) \"id\" not set")
		},
	},
	{
		name: "invalid --max-ttl type",
		args: []string{"update", "--id", "test-issuer", "--max-ttl", "invalid", "--data", `{"key":"value"}`},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "invalid argument \"invalid\" for \"-m, --max-ttl\" flag")
		},
	},
	{
		name: "invalid --data type",
		args: []string{"update", "--id", "test-issuer", "--max-ttl", "3000", "--data", "invalid-json"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "Failed to parse 'data' flag JSON:")
		},
	},
	{
		name:               "clientFactory returns an error",
		args:               []string{"update", "--id", "test-issuer", "--max-ttl", "3000", "--data", `{"key":"value"}`},
		clientFactoryError: errors.New("mock issuerFactory error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "mock issuerFactory error")
		},
	},
	{
		name:              "UpdateIssuer function returns an error",
		args:              []string{"update", "--id", "test-issuer", "--max-ttl", "3000", "--data", `{"key":"value"}`},
		updateIssuerError: errors.New("mock UpdateIssuer error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "mock UpdateIssuer error")
		},
	},
	{
		name:         "getStringFlagFunc returns an error for id key",
		args:         []string{"update", "--id", "test-issuer", "--max-ttl", "3000", "--data", `{"key":"value"}`},
		getStringErr: map[string]interface{}{"error": errors.New("mock getString error"), "key": "id"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "mock getString error")
		},
	},
	{
		name:         "getStringFlagFunc returns an error for data key",
		args:         []string{"update", "--id", "test-issuer", "--max-ttl", "3000", "--data", `{"key":"value"}`},
		getStringErr: map[string]interface{}{"error": errors.New("mock getString error"), "key": "data"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "mock getString error")
		},
	},
	{
		name:      "getIntFlagFunc returns an error",
		args:      []string{"update", "--id", "test-issuer", "--max-ttl", "3000", "--data", `{"key":"value"}`},
		getIntErr: errors.New("mock getInt error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "mock getInt error")
		},
	},
	{
		name: "issuer is updated",
		args: []string{"update", "--id", "test-issuer", "--max-ttl", "3000", "--data", `{"key":"value"}`},
		issuer: api.Issuer{
			ID:     "test-issuer",
			Type:   "aws",
			MaxTTL: 3000,
			Data: map[string]interface{}{
				"key": "value",
			},
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			output := "{\n  \"id\": \"test-issuer\",\n  \"type\": \"aws\",\n  \"max_ttl\": 3000,\n  \"data\": {\n    \"key\": \"value\"\n  }\n}\n"
			assert.Contains(t, stdout, output)
		},
	},
}

func TestUpdateIssuerCmd(t *testing.T) {
	for _, tc := range updateCmdTestCases {
		// Mock dependencies

		oldIntFunc := getIntFlagFunc
		oldStringFunc := getStringFlagFunc

		getIntFlagFunc = func(cmd *cobra.Command, key string) (int, error) {
			if tc.getIntErr != nil {
				return 0, tc.getIntErr
			}

			return cmd.Flags().GetInt(key)
		}
		getStringFlagFunc = func(cmd *cobra.Command, key string) (string, error) {
			if tc.getStringErr != nil && key == tc.getStringErr["key"] {
				return "", tc.getStringErr["error"].(error)
			}

			return cmd.Flags().GetString(key)
		}

		defer func() {
			getIntFlagFunc = oldIntFunc
			getStringFlagFunc = oldStringFunc
		}()
		t.Run(tc.name, func(t *testing.T) {
			cmd := newUpdateIssuerCmd(
				func(cmd *cobra.Command) (issuerClient, error) {
					return &mockIssuerClient{
						updateIssuer: func(issuerID string, issuerUpdate api.IssuerUpdate) (updated api.Issuer, err error) {
							return tc.issuer, tc.updateIssuerError
						},
					}, tc.clientFactoryError
				},
			)

			stdout, stderr, err := executeCommandForTest(t, cmd, tc.args...)
			tc.assert(t, stdout, stderr, err)
		})
	}
}

var deleteCmdTestCases = []struct {
	name               string
	getBoolErr         error
	getStringErr       error
	args               []string
	clientFactoryError error
	deleteIssuerError  error
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
		name: "missing --id flag",
		args: []string{"delete"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "required flag(s) \"id\" not set")
		},
	},
	{
		name:               "clientFactory returns an error",
		args:               []string{"delete", "--id", "test-issuer"},
		clientFactoryError: errors.New("mock issuerFactory error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "mock issuerFactory error")
		},
	},
	{
		name:         "getStringFlagFunc an error",
		args:         []string{"delete", "--id", "test-issuer"},
		getStringErr: errors.New("mock getString error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "mock getString error")
		},
	},
	{
		name:       "getBoolFunc an error",
		args:       []string{"delete", "--id", "test-issuer"},
		getBoolErr: errors.New("mock getBool error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "mock getBool error")
		},
	},
	{
		name:              "DeleteIssuer function returns an error",
		args:              []string{"delete", "--id", "test-issuer"},
		deleteIssuerError: errors.New("mock DeleteIssuer error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "mock DeleteIssuer error")
		},
	},
	{
		name: "issuer is deleted",
		args: []string{"delete", "--id", "test-issuer"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			output := "The issuer 'test-issuer' was deleted"
			assert.Contains(t, stdout, output)
		},
	},
}

func TestDeleteIssuerCmd(t *testing.T) {
	for _, tc := range deleteCmdTestCases {

		// Mock dependencies
		oldBoolFunc := getBoolFlagFunc
		oldStringFunc := getStringFlagFunc

		getBoolFlagFunc = func(cmd *cobra.Command, key string) (bool, error) {
			if tc.getBoolErr != nil {
				return false, tc.getBoolErr
			}

			return cmd.Flags().GetBool(key)
		}
		getStringFlagFunc = func(cmd *cobra.Command, key string) (string, error) {
			if tc.getStringErr != nil {
				return "", tc.getStringErr
			}

			return cmd.Flags().GetString(key)
		}
		defer func() {
			getBoolFlagFunc = oldBoolFunc
			getStringFlagFunc = oldStringFunc
		}()

		t.Run(tc.name, func(t *testing.T) {
			cmd := newDeleteIssuerCmd(
				func(cmd *cobra.Command) (issuerClient, error) {
					return &mockIssuerClient{
						deleteIssuer: func(issuerID string, keepSecrets bool) (err error) {
							return tc.deleteIssuerError
						},
					}, tc.clientFactoryError
				},
			)

			stdout, stderr, err := executeCommandForTest(t, cmd, tc.args...)
			tc.assert(t, stdout, stderr, err)
		})
	}
}

var listIssuerCmdTestCases = []struct {
	name               string
	args               []string
	getStringErr       error
	clientFactoryError error
	issurs             []api.Issuer
	ListIssuerError    error
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
		name:               "clientFactory returns an error",
		args:               []string{"list"},
		clientFactoryError: errors.New("mock issuerFactory error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "mock issuerFactory error")
		},
	},
	{
		name:            "List Issuer function returns an error",
		args:            []string{"list"},
		ListIssuerError: errors.New("mock ListIssuer error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "mock ListIssuer error")
		},
	},
	{
		name: "issuer are listed",
		args: []string{"list"},
		issurs: []api.Issuer{
			{
				ID: "test-issuer",
			},
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			output := "[\n  {\n    \"id\": \"test-issuer\",\n    \"type\": \"\",\n    \"max_ttl\": 0,\n    \"data\": null\n  }\n]\n"
			assert.Contains(t, stdout, output)
		},
	},
}

func TestListIssuerCmd(t *testing.T) {
	for _, tc := range listIssuerCmdTestCases {

		t.Run(tc.name, func(t *testing.T) {
			cmd := newListIssuersCmd(
				func(cmd *cobra.Command) (issuerClient, error) {
					return &mockIssuerClient{
						issuers: func() (issuers []api.Issuer, err error) {
							return tc.issurs, tc.ListIssuerError
						},
					}, tc.clientFactoryError
				},
			)

			stdout, stderr, err := executeCommandForTest(t, cmd, tc.args...)
			tc.assert(t, stdout, stderr, err)
		})
	}
}

var getIssuerCmdTestCases = []struct {
	name               string
	args               []string
	clientFactoryError error
	getStringErr       error
	issuer             api.Issuer
	GetIssuerError     error
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
		name: "missing --id flag",
		args: []string{"get"},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Contains(t, stderr, "required flag(s) \"id\" not set")
		},
	},
	{
		name:         "getStringFlagFunc an error",
		args:         []string{"get", "--id", "test-issuer"},
		getStringErr: errors.New("mock getString error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "mock getString error")
		},
	},
	{
		name:               "clientFactory returns an error",
		args:               []string{"get", "--id", "test-issuer"},
		clientFactoryError: errors.New("mock issuerFactory error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "mock issuerFactory error")
		},
	},
	{
		name:           "Get Issuer function returns an error",
		args:           []string{"get", "--id", "test-issuer"},
		GetIssuerError: errors.New("mock ListIssuer error"),
		assert: func(t *testing.T, stdout, stderr string, err error) {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "mock ListIssuer error")
		},
	},
	{
		name: "issuer retreived",
		args: []string{"get", "--id", "test-issuer"},
		issuer: api.Issuer{
			ID: "test-issuer",
		},
		assert: func(t *testing.T, stdout, stderr string, err error) {
			output := "{\n  \"id\": \"test-issuer\",\n  \"type\": \"\",\n  \"max_ttl\": 0,\n  \"data\": null\n}\n"
			assert.Contains(t, stdout, output)
		},
	},
}

func TestGetIssuerCmd(t *testing.T) {
	for _, tc := range getIssuerCmdTestCases {
		// Mock dependencies
		oldStringFunc := getStringFlagFunc

		getStringFlagFunc = func(cmd *cobra.Command, key string) (string, error) {
			if tc.getStringErr != nil {
				return "", tc.getStringErr
			}

			return cmd.Flags().GetString(key)
		}
		defer func() {
			getStringFlagFunc = oldStringFunc
		}()
		t.Run(tc.name, func(t *testing.T) {
			cmd := newGetIssuerCmd(
				func(cmd *cobra.Command) (issuerClient, error) {
					return &mockIssuerClient{
						issuer: func(issuerID string) (issuer api.Issuer, err error) {
							return tc.issuer, tc.GetIssuerError
						},
					}, tc.clientFactoryError
				},
			)

			stdout, stderr, err := executeCommandForTest(t, cmd, tc.args...)
			tc.assert(t, stdout, stderr, err)
		})
	}
}
