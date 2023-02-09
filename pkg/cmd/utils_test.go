package cmd

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"testing"
	"time"

	expect "github.com/Netflix/go-expect"
	"github.com/creack/pty"
	"github.com/hinshun/vt10x"
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
func executeCommandForTest(t *testing.T, cmd *cobra.Command, args ...string) (string, string, error) {
	t.Helper()

	rootCmd := newRootCommand()
	rootCmd.AddCommand(cmd)
	rootCmd.SetArgs(args)

	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	mockHelpText(cmd)

	rootCmd.SetOut(stdoutBuf)
	rootCmd.SetErr(stderrBuf)

	rootCmd.SetIn(bytes.NewReader(nil))
	err := cmd.Execute()

	// strip ansi from stdout and stderr because we're using promptui
	return stripAnsi(stdoutBuf.String()), stripAnsi(stderrBuf.String()), err
}

func executeCommandForTestWithPromptResponses(
	t *testing.T, cmd *cobra.Command, promptResponses []promptResponse,
) (stdOut string, stdErr string, cmdErr error) {

	// Mock short/long help output
	mockHelpText(cmd)

	if len(promptResponses) == 0 {
		return executeCmdWithBuffers(cmd)
	}

	consoleOutput := &bytes.Buffer{}
	c := setupVirtualConsole(t, consoleOutput)

	donec := make(chan struct{})
	go func() {
		defer close(donec)
		// Define virtual console interactivity.
		for _, p := range promptResponses {
			if p.prompt != "" {
				_, err := c.Expect(expect.String(p.prompt), expect.WithTimeout(500*time.Millisecond))
				if err != nil {
					// Exit the goroutine and allow the command execution to timeout
					t.Fatalf("Prompt not found.\nExpected:\n%s\nOutput:\n%s", p.prompt, stripAnsi(consoleOutput.String()))
					return
				}
			}
			if p.response != "" {
				c.SendLine(p.response)
			}
		}

		// Expect string that never gets output by command, this forces go-expect
		// to keep reading the command output to the buffer. Set the read timeout
		// to assume command has finished when there is no output for 0.5 seconds.
		c.Expect(expect.String("FINISHED"), expect.WithTimeout(500*time.Millisecond))
	}()

	executeCommandWithTimeOut(t, cmd)

	// Wait for virtual console to finish.
	<-donec

	return stripAnsi(consoleOutput.String()), stripAnsi(consoleOutput.String()), nil
}

func executeCommandWithTimeOut(t *testing.T, cmd *cobra.Command) {
	seconds := 3
	timeout := time.After(time.Duration(seconds) * time.Second)
	done := make(chan bool)
	go func() {
		// do your testing
		cmd.Execute()
		done <- true
	}()

	select {
	case <-timeout:
		t.Fatalf("Command execution timed out after %v seconds", seconds)
	case <-done:
	}
}

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

// stripAnsi is inspired by https://github.com/acarl005/stripansi
func stripAnsi(str string) string {
	return regexp.MustCompile(ansi).ReplaceAllString(str, "")
}

// mockHelpText recursively mocks the short and long help text of a command and its subcommands
func mockHelpText(cmd *cobra.Command) {
	cmd.Short = "HELP SHORT"
	cmd.Long = "HELP LONG"

	for _, subCmd := range cmd.Commands() {
		mockHelpText(subCmd)
	}
}

// executeCmdWithBuffers executes the command as-is and returns stdout, stderr and error
func executeCmdWithBuffers(cmd *cobra.Command) (stdOut string, stdErr string, cmdErr error) {
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	cmd.SetOut(stdoutBuf)
	cmd.SetErr(stderrBuf)

	err := cmd.Execute()

	return stripAnsi(stdoutBuf.String()), stripAnsi(stderrBuf.String()), err
}

// setupVirtualConsole will create a virtual console that for use in tests.
// The virtual console will override the stdout, stderr, and stdin, and reset
// them at the end of the test. As stdout, and stderr are overriden it can be
// difficult to debug a test as standard print statements will not be visible.
// To fix this issue use the verbose option, -v, with go test.
func setupVirtualConsole(t *testing.T, w io.Writer) *expect.Console {
	t.Helper()
	c, err := newConsole(expect.WithStdout(w))
	if err != nil {
		t.Fatalf("unable to create virtual console: %v", err)
	}
	origIn := os.Stdin
	origOut := os.Stdout
	origErr := os.Stderr
	t.Cleanup(func() {
		os.Stdin = origIn
		os.Stdout = origOut
		os.Stderr = origErr
		c.Close()
	})
	os.Stdin = c.Tty()
	os.Stdout = c.Tty()
	os.Stderr = c.Tty()
	return c
}

// newConsole returns a new expect.Console that multiplexes the
// Stdin/Stdout to a VT10X terminal, allowing Console to interact with an
// application sending ANSI escape sequences.
func newConsole(opts ...expect.ConsoleOpt) (*expect.Console, error) {
	ptm, pts, err := pty.Open()
	if err != nil {
		return nil, err
	}
	term := vt10x.New(vt10x.WithWriter(pts))
	c, err := expect.NewConsole(append(opts, expect.WithStdin(ptm), expect.WithStdout(term), expect.WithCloser(pts, ptm))...)
	if err != nil {
		return nil, err
	}
	return c, nil
}
