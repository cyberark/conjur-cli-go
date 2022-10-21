package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// promptResponse is a prompt-response pair representing an expected prompt from stdout and
// the response to write on stdin when such a prompt is encountered
type promptResponse struct {
	prompt   string // optional, because
	response string
}

// executeCommandForTest executes a cobra command in-memory and returns stdout, stderr and error
func executeCommandForTest(t *testing.T, c *cobra.Command, args ...string) (string, string, error) {
	return executeCommandForTestWithPromptResponses(t, c, nil, args...)
}

// executeCommandForTestWithPromptResponses executes a cobra command in-memory and returns stdout, stderr and error.
// It takes a as input a slice representing a sequence of prompt-responses, each containing an expected
// prompt from stdout and the response to write on stdin when such a prompt is encountered.
func executeCommandForTestWithPromptResponses(
	t *testing.T, c *cobra.Command, promptResponses []promptResponse, args ...string,
) (string, string, error) {
	t.Helper()

	cmd := newRootCommand()
	cmd.AddCommand(c)
	cmd.SetArgs(args)

	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	cmd.SetOut(stdoutBuf)
	cmd.SetErr(stderrBuf)

	cmdErr, stdinErr := executeCommandWithPromptResponses(
		cmd,
		promptResponses,
	)

	if stdinErr != nil {
		t.Error(stdinErr)
	}

	// strip ansi from stdout and stderr because we're using promptui
	return stripAnsi(stdoutBuf.String()), stripAnsi(stderrBuf.String()), cmdErr
}

func executeCommandWithPromptResponses(
	cmd *cobra.Command, promptResponses []promptResponse,
) (cmdErr error, stdinErr error) {
	if len(promptResponses) == 0 {
		return cmd.Execute(), nil
	}

	endSignal := make(chan struct{}, 1)
	doneSignal := make(chan error, 1)

	stdoutBuf := new(bytes.Buffer)
	cmd.SetOut(io.MultiWriter(cmd.OutOrStdout(), stdoutBuf))

	stdinReader, stdinWriter := io.Pipe()
	cmd.SetIn(io.NopCloser(stdinReader))

	go func() {
		defer func() {
			stdinReader.Close()
			stdinWriter.Close()
		}()

		var goRoutineErr error
		reader := bufio.NewReader(stdoutBuf)
		promptResponsesIndex := 0

		timer := time.NewTimer(1 * time.Second)
		defer timer.Stop()

		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

	L:
		for promptResponsesIndex < len(promptResponses) {
			select {
			case <-timer.C:
				goRoutineErr = fmt.Errorf(
					"a whole second passed without the command issuing the expected prompt: %q",
					promptResponses[promptResponsesIndex].prompt,
				)

				break L
			case <-endSignal:
				break L
			case <-ticker.C:
				line, err := reader.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						// Sleep is important to avoid a busy goroutine
						// that never yields.
						continue
					}

					goRoutineErr = err
					break L
				}

				// Strip ANSI because it's just noise
				line = stripAnsi(line)

				// Try to match promptResponses in order
				promptResponse := promptResponses[promptResponsesIndex]
				if strings.Contains(line, promptResponse.prompt) {
					promptResponsesIndex++

					stdinWriter.Write([]byte(promptResponse.response + "\n"))
				}
			}
		}

		if goRoutineErr == nil && promptResponsesIndex < len(promptResponses) {
			goRoutineErr = fmt.Errorf(
				"promptResponses not exhausted, remaining: %+v",
				promptResponses[promptResponsesIndex:],
			)
		}
		doneSignal <- goRoutineErr
	}()

	err := cmd.Execute()

	endSignal <- struct{}{}

	if goroutineErr := <-doneSignal; goroutineErr != nil {
		return err, goroutineErr
	}

	return err, nil
}

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

// stripAnsi is inspired by https://github.com/acarl005/stripansi
func stripAnsi(str string) string {
	return regexp.MustCompile(ansi).ReplaceAllString(str, "")
}
