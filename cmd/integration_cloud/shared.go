//go:build integration
// +build integration

package main

import (
	"bytes"
	"github.com/cyberark/conjur-api-go/conjurapi"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	pathToBinary = "conjur"
)

func newConjurTestCLI(t *testing.T) (cli *testConjurCLI) {
	homeDir := t.TempDir()

	cli = &testConjurCLI{
		homeDir: homeDir,
	}

	return
}

type testConjurCLI struct {
	homeDir string
}

func (cli *testConjurCLI) InitAndLoginAsAdmin(t *testing.T) {
	// Initialize the CLI
	cli.Init(t)

	// Login as admin
	cli.LoginAsAdmin(t)
}

func (cli *testConjurCLI) Init(t *testing.T) {
	stdOut, _, err := cli.RunWithYes("init", string(conjurapi.EnvironmentCC), "-u", os.Getenv("CONJUR_APPLIANCE_URL"), "--force-netrc", "--force")
	assertInitCmd(t, err, stdOut, cli.homeDir)
}

func (cli *testConjurCLI) LoginAsAdmin(t *testing.T) {
	stdOut, stdErr, err := cli.Run("login", "-i", os.Getenv("IDENTITY_USERNAME_CLOUD"), "-p", os.Getenv("IDENTITY_PASSWORD_CLOUD"))
	assertLoginCmd(t, err, stdOut, stdErr)
}

func (cli *testConjurCLI) RunWithYes(args ...string) (stdOut string, stdErr string, err error) {
	return cli.RunWithStdin(strings.NewReader("yes\n"), args...)
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

func assertInitCmd(t *testing.T, err error, stdOut string, homeDir string) {
	assert.NoError(t, err)
	assert.Contains(t, stdOut, "Wrote certificate to "+homeDir+"/conjur-server.pem")
	assert.Contains(t, stdOut, "Wrote configuration to "+homeDir+"/.conjurrc")
}

func assertLoginCmd(t *testing.T, err error, stdOut string, stdErr string) {
	assert.NoError(t, err)
	assert.Contains(t, stdOut, "Logged in\n")
	assert.Equal(t, "", stdErr)
}
