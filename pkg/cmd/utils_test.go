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

	return stripAnsi(stdoutBuf.String()), stripAnsi(stderrBuf.String()), err
}

func executeCommandForTestWithPipeResponses(
	t *testing.T, cmd *cobra.Command, content string,
) (stdOut string, cmdErr error) {
	t.Helper()

	mockHelpText(cmd)

	// Send CLI stdOut and stdErr to Pipes and consolidate
	outR, outW, _ := os.Pipe()
	errR, errW, _ := os.Pipe()
	combinedOutReader := io.MultiReader(outR, errR)

	defer func(origOut *os.File) { os.Stdout = origOut }(os.Stdout)
	defer func(origErr *os.File) { os.Stderr = origErr }(os.Stderr)
	os.Stdout = outW
	os.Stderr = errW

	// Receive CLI stdIn from Pipe
	inR, inW, _ := os.Pipe()
	_, err := inW.Write([]byte(content))
	if err != nil {
		t.Fatal(err)
	}
	inW.Close()

	defer func(origIn *os.File) { os.Stdin = origIn }(os.Stdin)
	os.Stdin = inR

	// Run CLI cmd
	err = cmd.Execute()

	// Collect stdOut
	outW.Close()
	errW.Close()
	out, outErr := io.ReadAll(combinedOutReader)
	if outErr != nil {
		t.Fatal(outErr)
	}

	return stripAnsi(string(out)), err
}

func executeCommandForTestWithPromptResponses(
	t *testing.T, cmd *cobra.Command, promptResponses []promptResponse,
) (stdOut string, cmdErr error) {

	// Mock short/long help output
	mockHelpText(cmd)

	consoleOutput := &bytes.Buffer{}
	c := setupVirtualConsole(t, consoleOutput)
	defer c.Close()

	donec := make(chan struct{})
	go func() {
		defer close(donec)
		// Define virtual console interactivity.
		for _, p := range promptResponses {
			if p.prompt != "" {
				_, _ = c.ExpectString(p.prompt)
			}
			if p.response != "" {
				_, _ = c.SendLine(p.response)
			}
		}

		// Wait until EOF gets printed to the virtual console, which happens in the code below
		// when the command finishes executing. This forces the virtual console to
		// continue reading input until the command finishes. Set a timeout
		// to ensure that the virtual console doesn't wait forever.
		_, err := c.Expect(expect.String("TEST_COMMAND_EXITED"), expect.WithTimeout(10*time.Second))
		// If the timeout is reached, we should fail the test.
		assert.NoError(t, err)
	}()

	// Execute command with args.
	err := cmd.Execute()
	// Print EOF to virtual console to signal that the command has finished executing.
	cmd.Print("TEST_COMMAND_EXITED")
	// Wait for virtual console to finish.
	<-donec

	return stripAnsi(consoleOutput.String()), err
}

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

// stripAnsi is inspired by https://github.com/acarl005/stripansi
func stripAnsi(str string) string {
	return regexp.MustCompile(ansi).ReplaceAllString(str, "")
}

func mockHelpText(cmd *cobra.Command) {
	cmd.Short = "HELP SHORT"
	cmd.Long = "HELP LONG"

	for _, subCmd := range cmd.Commands() {
		mockHelpText(subCmd)
	}
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
