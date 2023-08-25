//go:build integration
// +build integration

package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const pathToBinary = "conjur"

const insecureModeWarning = "Warning: Running the command with '--insecure' makes your system vulnerable to security attacks\n" +
	"If you prefer to communicate with the server securely you must reinitialize the client in secure mode.\n"
const selfSignedWarning = "Warning: Using self-signed certificates is not recommended and could lead to exposure of sensitive data\n"

const testPolicy = `
- !variable meow
- !variable woof
- !user alice
- !host bob

- !permit
  resource: !variable meow
  role: !user alice
  privileges: [ read ]
`

func newConjurTestCLI(t *testing.T) (cli *testConjurCLI) {
	homeDir := t.TempDir()
	// Lean on the uniqueness of temp directories
	account := strings.Replace(homeDir, "/", "", -1)

	cli = &testConjurCLI{
		homeDir: homeDir,
		account: account,
	}

	// Create a Conjur account
	cleanUpConjurAccount := prepareConjurAccount(account)
	// Clean up the Conjur account after the tests complete
	t.Cleanup(cleanUpConjurAccount)

	return
}

type testConjurCLI struct {
	homeDir string
	account string
}

func (cli *testConjurCLI) InitAndLoginAsAdmin(t *testing.T) {
	// Initialize the CLI
	cli.Init(t)

	// Login as admin
	cli.LoginAsAdmin(t)

	// Load test policy
	cli.LoadPolicy(t, testPolicy)
}

func (cli *testConjurCLI) Init(t *testing.T) {
	stdOut, _, err := cli.Run("init", "-a", cli.account, "-u", "http://conjur", "-i", "--force-netrc", "--force")
	assertInitCmd(t, err, stdOut, cli.homeDir)
}

func (cli *testConjurCLI) InitWithTrailingSlash(t *testing.T) {
	stdOut, _, err := cli.Run("init", "-a", cli.account, "-u", "http://conjur/", "-i", "--force-netrc", "--force")
	assertInitCmd(t, err, stdOut, cli.homeDir)
}

func (cli *testConjurCLI) LoginAsAdmin(t *testing.T) {
	stdOut, stdErr, err := cli.Run("login", "-i", "admin", "-p", makeDevRequest("retrieve_api_key", map[string]string{"role_id": cli.account + ":user:admin"}))
	assertLoginCmd(t, err, stdOut, stdErr)
}

func (cli *testConjurCLI) LoginAsHost(t *testing.T, host string) {
	stdOut, stdErr, err := cli.Run("login", "-i", "host/"+host, "-p", makeDevRequest("retrieve_api_key", map[string]string{"role_id": cli.account + ":host:" + host}))
	assertLoginCmd(t, err, stdOut, stdErr)
}

func (cli *testConjurCLI) LoadPolicy(t *testing.T, policyText string) {
	stdOut, stdErr, err := cli.RunWithStdin(
		bytes.NewReader([]byte(policyText)),
		"policy", "load", "-b", "root", "-f", "-",
	)
	assertPolicyLoadCmd(t, err, stdOut, stdErr)
}

func (cli *testConjurCLI) RunWithStdin(stdIn io.Reader, args ...string) (stdOut string, stdErr string, err error) {
	cmd := exec.Command(pathToBinary, args...)
	stdOutBuffer := new(bytes.Buffer)
	stdErrBuffer := new(bytes.Buffer)
	cmd.Stdin = stdIn
	cmd.Stdout = io.MultiWriter(stdOutBuffer, os.Stdout)
	cmd.Stderr = io.MultiWriter(stdErrBuffer, os.Stderr)

	cmd.Env = append(cmd.Env, "HOME="+cli.homeDir)

	err = cmd.Run()
	return stdOutBuffer.String(), stdErrBuffer.String(), err
}

func (cli *testConjurCLI) Run(args ...string) (stdOut string, stdErr string, err error) {
	return cli.RunWithStdin(nil, args...)
}

func makeDevRequest(action string, params map[string]string) string {
	url, _ := url.Parse("http://conjur/dev")
	query := url.Query()
	for k, v := range params {
		query.Add(k, v)
	}
	query.Add("action", action)

	url.RawQuery = query.Encode()

	resp, err := http.Get(url.String())
	if err != nil {
		log.Fatalln(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	return string(body)
}

func prepareConjurAccount(account string) func() {
	makeDevRequest("destroy_account", map[string]string{"id": account})
	makeDevRequest("create_account", map[string]string{"id": account})
	return func() {
		makeDevRequest("destroy_account", map[string]string{"id": account})
	}
}

func loadPolicy(account string, policy string) {
	makeDevRequest("load_policy", map[string]string{"resource_id": account + ":policy:root", "policy": policy})
}

func loadPolicyFile(account, policyFile string) {
	policy, err := os.ReadFile(policyFile)
	if err != nil {
		panic(err)
	}
	loadPolicy(account, string(policy))
}

func createSecret(account string, variable string, value string) {
	makeDevRequest("create_secret", map[string]string{"resource_id": account + ":variable:" + variable, "value": value})
}

func assertInitCmd(t *testing.T, err error, stdOut string, homeDir string) {
	assert.NoError(t, err)
	assert.Contains(t, stdOut, "Wrote configuration to "+homeDir+"/.conjurrc\n")
}

func assertLoginCmd(t *testing.T, err error, stdOut string, stdErr string) {
	assert.NoError(t, err)
	assert.Contains(t, stdOut, "Logged in\n")
	assert.Equal(t, "", stdErr)
}

func assertWhoamiCmd(t *testing.T, err error, stdOut string, stdErr string) {
	assert.NoError(t, err)
	assert.Contains(t, stdOut, "token_issued_at")
	assert.Contains(t, stdOut, "client_ip")
	assert.Contains(t, stdOut, "user_agent")
	assert.Contains(t, stdOut, "account")
	assert.Contains(t, stdOut, "username")
	assert.Equal(t, "", stdErr)
}

func assertAuthenticateCmd(t *testing.T, err error, stdOut string, stdErr string) {
	assert.NoError(t, err)
	assert.Contains(t, stdOut, "protected")
	assert.Contains(t, stdOut, "payload")
	assert.Contains(t, stdOut, "signature")
	assert.Equal(t, "", stdErr)
}

func assertNotFound(t *testing.T, err error, stdOut string, stdErr string) {
	assert.Error(t, err)
	assert.Equal(t, "", stdOut)
	assert.Contains(t, stdErr, "Error: 404 Not Found.")
}

func assertPolicyLoadCmd(t *testing.T, err error, stdOut string, stdErr string) {
	assert.NoError(t, err)
	assert.Contains(t, stdOut, "created_roles")
	assert.Contains(t, stdOut, "version")
	assert.Equal(t, "Loaded policy 'root'\n", stdErr)
}

func assertSetVariableCmd(t *testing.T, err error, stdOut string, stdErr string) {
	assert.NoError(t, err)
	assert.Equal(t, "Value added\n", stdOut)
	assert.Equal(t, "", stdErr)
}

func assertGetVariableCmd(t *testing.T, err error, stdOut string, stdErr string, excpectedValue string) {
	assert.NoError(t, err)
	assert.Equal(t, excpectedValue+"\n", stdOut)
	assert.Equal(t, "", stdErr)
}

func assertGetTwoVariablesCmd(t *testing.T, err error, stdOut string, stdErr string) {
	assert.NoError(t, err)
	assert.Contains(t, stdOut, "moo")
	assert.Contains(t, stdOut, "quack")
	assert.Equal(t, "", stdErr)
}

func assertExistsCmd(t *testing.T, err error, stdOut string, stdErr string) {
	assert.NoError(t, err)
	assert.Equal(t, "false\n", stdOut)
}

func assertLogoutCmd(t *testing.T, err error, stdOut string, stdErr string) {
	assert.NoError(t, err)
	assert.Contains(t, stdOut, "Logged out\n")
	assert.Equal(t, "", stdErr)
}

func assertAPIKeyRotationCmd(t *testing.T, err error, stdOut string, stdErr string, priorAPIKey string) {
	assert.NoError(t, err)
	assert.Regexp(t, regexp.MustCompile("[a-zA-Z0-9]{45,60}\n"), stdOut)
	assert.NotEqual(t, priorAPIKey, stdOut)
	assert.Equal(t, "", stdErr)
}
