//go:build integration
// +build integration

package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const pathToBinary = "conjur"

func runCLIWithStdin(stdIn io.Reader, home string, args ...string) (stdOut string, stdErr string, err error) {
	cmd := exec.Command(pathToBinary, args...)
	stdOutBuffer := new(bytes.Buffer)
	stdErrBuffer := new(bytes.Buffer)
	cmd.Stdin = stdIn
	cmd.Stdout = io.MultiWriter(stdOutBuffer, os.Stdout)
	cmd.Stderr = io.MultiWriter(stdErrBuffer, os.Stderr)

	cmd.Env = append(cmd.Env, "HOME="+home)

	err = cmd.Run()
	return stdOutBuffer.String(), stdErrBuffer.String(), err
}

func runCLI(home string, args ...string) (stdOut string, stdErr string, err error) {
	return runCLIWithStdin(nil, home, args...)
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	return string(body)
}

func TestIntegration(t *testing.T) {
	var (
		stdOut string
		stdErr string
		err    error
	)
	tmpDir := t.TempDir()
	// Lean on the uniqueness of temp directories
	id := strings.Replace(tmpDir, "/", "", -1)

	account := id

	t.Run("ensure binary exists", func(t *testing.T) {
		_, err = exec.LookPath(pathToBinary)
		assert.NoError(t, err)
	})

	// Setup brand new account in Conjur
	makeDevRequest("purge", nil)
	makeDevRequest("destroy_account", map[string]string{"id": account})
	makeDevRequest("create_account", map[string]string{"id": account})
	defer makeDevRequest("destroy_account", map[string]string{"id": account})

	t.Run("whoami before init", func(t *testing.T) {
		stdOut, stdErr, err = runCLI(tmpDir, "whoami")
		assert.Error(t, err)
		assert.Equal(t, "", stdOut)
		assert.Equal(t, "Error: Missing required configuration for Conjur API URL\n", stdErr)
	})

	t.Run("init", func(t *testing.T) {
		stdOut, stdErr, err = runCLI(tmpDir, "init", "-a", account, "-u", "http://conjur", "--force")
		assert.NoError(t, err)
		assert.Equal(t, "Wrote configuration to "+tmpDir+"/.conjurrc\n", stdOut)
		assert.Equal(t, "", stdErr)
	})

	t.Run("login", func(t *testing.T) {
		stdOut, stdErr, err = runCLI(tmpDir, "login", "-u", "admin", "-p", makeDevRequest("retrieve_api_key", map[string]string{"role_id": account + ":user:admin"}))
		assert.NoError(t, err)
		assert.Equal(t, "Logged in\n", stdOut)
		assert.Equal(t, "", stdErr)
	})

	t.Run("whoami after init", func(t *testing.T) {
		stdOut, stdErr, err = runCLI(tmpDir, "whoami")
		assert.NoError(t, err)
		assert.Contains(t, stdOut, "token_issued_at")
		assert.Contains(t, stdOut, "client_ip")
		assert.Contains(t, stdOut, "user_agent")
		assert.Contains(t, stdOut, "account")
		assert.Contains(t, stdOut, "username")
		assert.Equal(t, "", stdErr)
	})

	t.Run("get variable before policy load", func(t *testing.T) {
		stdOut, stdErr, err = runCLI(tmpDir, "variable", "get", "-i", "meow")
		assert.Error(t, err)
		assert.Equal(t, "", stdOut)
		assert.Contains(t, stdErr, "Error: 404 Not Found.")
	})

	t.Run("policy load", func(t *testing.T) {
		stdOut, stdErr, err = runCLIWithStdin(
			bytes.NewReader([]byte("- !variable meow")),
			tmpDir,
			"policy", "load", "-b", "root", "-f", "-",
		)
		assert.NoError(t, err)
		assert.Contains(t, stdOut, "created_roles")
		assert.Contains(t, stdOut, "version")
		assert.Equal(t, "Loaded policy 'root'\n", stdErr)
	})

	t.Run("set variable after policy load", func(t *testing.T) {
		stdOut, stdErr, err = runCLI(tmpDir, "variable", "set", "-i", "meow", "-v", "moo")
		assert.NoError(t, err)
		assert.Equal(t, "Value added\n", stdOut)
		assert.Equal(t, "", stdErr)
	})

	t.Run("get variable after policy load", func(t *testing.T) {
		stdOut, stdErr, err = runCLI(tmpDir, "variable", "get", "-i", "meow")
		assert.NoError(t, err)
		assert.Equal(t, "moo\n", stdOut)
		assert.Equal(t, "", stdErr)
	})
}
