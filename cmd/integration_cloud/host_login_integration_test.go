//go:build integration
// +build integration

package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testHostID        = "host/data/my-host"
	testHostShortName = "data/my-host"
	testHostName      = "my-host"
	testPolicyBranch  = "data"
	testPolicyFile    = "host_login_policy.yml"
)

type policyLoadResponse struct {
	CreatedRoles map[string]struct {
		Id     string `json:"id"`
		ApiKey string `json:"api_key"`
	} `json:"created_roles"`
	Version int `json:"version"`
}

func TestHostLoginIntegration(t *testing.T) {
	cli := newConjurTestCLI(t)
	cli.InitAndLoginAsAdmin(t)

	t.Run("host", func(t *testing.T) {
		apiKey := ensureHostExists(t, cli)
		assert.NotEmpty(t, apiKey, "API key should not be empty")

		loginAsHost(t, cli, apiKey)
		verifyHostIdentity(t, cli)
	})
}

func ensureHostExists(t *testing.T, cli *testConjurCLI) string {
	stdOut, stdErr, err := cli.Run("policy", "load", "-b", testPolicyBranch, "-f", testPolicyFile)
	assert.NoError(t, err, "Failed to load host login policy")
	assert.Contains(t, stdErr, "Loaded policy 'data'")

	var response policyLoadResponse
	err = json.Unmarshal([]byte(stdOut), &response)
	assert.NoError(t, err, "Failed to unmarshal policy load response")

	switch len(response.CreatedRoles) {
	case 1:
		return extractAPIKey(response)
	case 0:
		return rotateHostAPIKey(t, cli)
	default:
		t.Fatalf("Expected 0 or 1 created roles, got %d: %v", len(response.CreatedRoles), response)
		return ""
	}
}

func extractAPIKey(response policyLoadResponse) string {
	for _, role := range response.CreatedRoles {
		return role.ApiKey
	}
	return ""
}

func rotateHostAPIKey(t *testing.T, cli *testConjurCLI) string {
	stdOut, _, err := cli.Run("host", "rotate-api-key", "-i", testHostShortName)
	assert.NoError(t, err, "Failed to rotate host API key")

	return strings.TrimSpace(stdOut)
}

func loginAsHost(t *testing.T, cli *testConjurCLI, apiKey string) {
	stdOut, _, err := cli.Run("-d", "login", "-i", testHostID, "-p", apiKey)
	assert.NoError(t, err, "Failed to login as host")
	assert.Contains(t, stdOut, "Logged in", "Unexpected login output")
}

func verifyHostIdentity(t *testing.T, cli *testConjurCLI) {
	stdOut, _, err := cli.Run("whoami")
	assert.NoError(t, err, "Failed to run whoami")
	assert.Contains(t, stdOut, testHostName, "Unexpected whoami output")
}
