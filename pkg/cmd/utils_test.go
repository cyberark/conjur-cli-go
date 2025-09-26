package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// promptResponse is a prompt-response pair representing an expected prompt from stdout and
// the response to write on stdin when such a prompt is encountered
type promptResponse struct {
	// Optional. If empty, the response is sent to stdin without waiting for a prompt
	prompt string

	// A new line character is added to the end of the response.
	// This makes it so that an empty response is equivalent to sending a newline to stdin.
	response string
}

// executeCommandForTest executes a cobra command in-memory and returns stdout, stderr and error
func executeCommandForTest(t *testing.T, c *cobra.Command, args ...string) (string, string, error) {
	t.Helper()

	cmd := newRootCommand()
	cmd.AddCommand(c)
	cmd.SetArgs(args)

	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	mockHelpText(cmd)

	cmd.SetOut(stdoutBuf)
	cmd.SetErr(stderrBuf)

	cmd.SetIn(bytes.NewReader(nil))
	err := cmd.Execute()

	return stdoutBuf.String(), stderrBuf.String(), err
}

func executeCommandForTestWithPipeResponses(
	t *testing.T, cmd *cobra.Command, content string,
) (stdOut string, cmdErr error) {
	t.Helper()

	// Mock short/long help output
	mockHelpText(cmd)

	cleanup, in, out, done := mockStdio(t)

	_, err := in.Write([]byte(content))
	// Run CLI cmd
	err = cmd.Execute()

	cleanup()
	<-done
	return out.String(), err
}

func executeCommandForTestWithPromptResponses(
	t *testing.T, cmd *cobra.Command, promptResponses []promptResponse,
) (stdOut string, cmdErr error) {
	t.Helper()

	// Mock short/long help output
	mockHelpText(cmd)

	cleanup, in, out, done := mockStdio(t)

	go func() {
		for _, pr := range promptResponses {
			for range time.Tick(100 * time.Millisecond) {
				if strings.Contains(out.String(), pr.prompt) {
					if len(pr.response) > 0 {
						_, err := in.Write([]byte(pr.response + "\n"))
						assert.NoError(t, err)
					}
					break
				}
			}
		}
	}()

	err := cmd.Execute()

	cleanup()
	<-done
	return out.String(), err
}

func mockStdio(t *testing.T) (func(), io.Writer, *bytes.Buffer, chan bool) {
	t.Helper()

	// Create pipes for stdin and stdout
	stdinR, stdinW, err := os.Pipe()
	assert.NoError(t, err)
	stdoutR, stdoutW, err := os.Pipe()
	assert.NoError(t, err)

	// Save original stdin, stdout, and stderr
	origStdin, origStdout, origStderr := os.Stdin, os.Stdout, os.Stderr

	// Redirect stdin, stdout, and stderr
	os.Stdin, os.Stdout, os.Stderr = stdinR, stdoutW, stdoutW

	cleanup := func() {
		os.Stdin, os.Stdout, os.Stderr = origStdin, origStdout, origStderr
		_, _, _ = stdinR.Close(), stdinW.Close(), stdoutW.Close()
	}

	done := make(chan bool)
	output := &bytes.Buffer{}
	go func() {
		defer close(done)
		_, err = io.Copy(output, stdoutR)
		stdoutR.Close()
	}()

	return cleanup, stdinW, output, done
}

func mockHelpText(cmd *cobra.Command) {
	cmd.Short = "HELP SHORT"
	cmd.Long = "HELP LONG"

	for _, subCmd := range cmd.Commands() {
		mockHelpText(subCmd)
	}
}
