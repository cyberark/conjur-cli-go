package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
)

const pathToBinary = "conjur"

func newConjurCLI(homeDir string) *conjurCLI {
	return &conjurCLI{
		homeDir: homeDir,
	}
}

type conjurCLI struct {
	homeDir string
}

func (cli *conjurCLI) RunWithStdin(stdIn io.Reader, args ...string) (stdOut string, stdErr string, err error) {
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

func (cli *conjurCLI) Run(args ...string) (stdOut string, stdErr string, err error) {
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

func createSecret(account string, variable string, value string) {
	makeDevRequest("create_secret", map[string]string{"resource_id": account + ":variable:" + variable, "value": value})
}
