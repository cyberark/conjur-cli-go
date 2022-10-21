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
	response string // A new line character is added to the end of the response,
	// unless if the last character of the response is already a new line.
	// This makes it so that an empty response is equivalent to sending a newline to stdin.
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

func maybeAddNewLineSuffix(s string) string {
	if strings.HasSuffix(s, "\n") {
		return s
	}

	return s + "\n"
}

const maxWaitTimeForPrompt = 2 * time.Second
const promptCheckInterval = 10 * time.Millisecond

func executeCommandWithPromptResponses(
	cmd *cobra.Command, promptResponses []promptResponse,
) (cmdErr error, stdinErr error) {
	// For no prompt-responses we simply execute with an empty stdin :)
	if len(promptResponses) == 0 {
		cmd.SetIn(bytes.NewReader(nil))
		return cmd.Execute(), nil
	}

	// For prompt-responses we create a pipe for stdin and use goroutine to carry out the
	// required writes to stdin in response to prompts
	endSignal := make(chan struct{}, 1)
	doneSignal := make(chan error, 1)

	stdoutBuf := new(bytes.Buffer)
	cmd.SetOut(io.MultiWriter(cmd.OutOrStdout(), stdoutBuf))

	stdinReader, stdinWriter := io.Pipe()
	// Close the reader here so that this happens after the command invocation
	defer func() {
		stdinReader.Close()
	}()
	cmd.SetIn(io.NopCloser(stdinReader))

	// This goroutine schedules writes to the write-end of the stdin pipe based on the prompt-responses
	go func() {
		// Close the writer here in the goroutine to signal to the readers that we're done writing after
		// what's left in the buffer
		defer func() {
			stdinWriter.Close()
		}()

		var goRoutineErr error
		reader := bufio.NewReader(stdoutBuf)
		promptResponsesIndex := 0

		timer := time.NewTimer(maxWaitTimeForPrompt)
		defer timer.Stop()

		ticker := time.NewTicker(promptCheckInterval)
		defer ticker.Stop()

	L:
		for promptResponsesIndex < len(promptResponses) {
			promptResponse := promptResponses[promptResponsesIndex]
			// Empty prompt means immediately write to stdin without waiting for a stdout prompt
			// This is useful when a command is reading from stdin such as policy loading.
			// TODO: what happens when you have a sequence of empty prompts ?
			if promptResponse.prompt == "" {
				promptResponsesIndex++

				stdinWriter.Write([]byte(maybeAddNewLineSuffix(promptResponse.response)))
				continue
			}

			select {
			// This case makes sure that you don't wait forever for a prompt that never shows up.
			// Here we timebox the wait on a prompt.
			// Additionally ensures we don't leak this goroutine because it ensures that after
			// a second of waiting we break the circuit.
			case <-timer.C:
				goRoutineErr = fmt.Errorf(
					"maxWaitTimeForPrompt was exceeded without the command issuing the expected prompt: %q",
					promptResponse.prompt,
				)
				break L
			// This cases makes it so that you don't actually have to wait for the timebox above
			// to be exhausted all if the command is already done, it's a cleaner signal.
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
				if strings.Contains(line, promptResponse.prompt) {
					promptResponsesIndex++

					stdinWriter.Write([]byte(maybeAddNewLineSuffix(promptResponse.response)))
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
