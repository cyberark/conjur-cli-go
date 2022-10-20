package cmd

import (
	"bytes"
	"io"
	"regexp"
	"testing"

	"github.com/spf13/cobra"
)

// executeCommandForTest executes a cobra command in-memory and returns stdout, stderr and error
func executeCommandForTest(t *testing.T, c *cobra.Command, args ...string) (string, string, error) {
	return executeCommandForTestWithStdin(t, c, "", args...)
}

// stdinBuffer is to address the fact that promptui is greedy and attempts to consume all the bytes
// from the stdin reader when it runs a prompt. For testing we want to provide a special stdin reader
// that will return butes up to the length of the destination or up to the next newline charater.
// This allows us to provide responses to multiple prompts for testing.
type stdinBuffer struct {
	bytes.Buffer
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (b *stdinBuffer) Read(dst []byte) (int, error) {
	if b.Buffer.Len() == 0 {
		return 0, io.EOF
	}

	nextNewlineIndex := bytes.IndexByte(b.Buffer.Bytes(), byte('\n'))
	if nextNewlineIndex == -1 {
		nextNewlineIndex = b.Buffer.Len() - 1
	}

	n := min(nextNewlineIndex+1, len(dst))

	copy(dst, b.Buffer.Next(n))
	return n, nil
}

// executeCommandForTestWithStdin executes a cobra command in-memory and returns stdout, stderr and error. It takes a
// as input a string representing stdin, and sets a custom stdin buffer on the command. When the custom stdin buffer
// is Read is will return bytes up to the length of the destination or up to the next newline charater, see
// StdinBuffer for details.
func executeCommandForTestWithStdin(t *testing.T, c *cobra.Command, stdin string, args ...string) (string, string, error) {
	t.Helper()

	rootCmd := newRootCommand()
	rootCmd.AddCommand(c)

	stdinBuf := new(stdinBuffer)
	stdinBuf.WriteString(stdin)

	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	rootCmd.SetOut(stdoutBuf)
	rootCmd.SetErr(stderrBuf)
	rootCmd.SetIn(stdinBuf)

	rootCmd.SetArgs(args)

	err := rootCmd.Execute()

	// strip ansi from stdout and stderr because we're using promptui
	return stripAnsi(stdoutBuf.String()), stripAnsi(stderrBuf.String()), err
}

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

// stripAnsi is inspired by https://github.com/acarl005/stripansi
func stripAnsi(str string) string {
	return regexp.MustCompile(ansi).ReplaceAllString(str, "")
}
