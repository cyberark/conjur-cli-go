package utils

import (
	"bytes"
	"testing"

	"github.com/chzyer/readline"
	"github.com/stretchr/testify/assert"
)

func TestWrite(t *testing.T) {
	testCases := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "empty",
			input:    []byte{},
			expected: []byte{},
		},
		{
			name:     "simple",
			input:    []byte("foo"),
			expected: []byte("foo"),
		},
		{
			name:     "bell",
			input:    []byte{readline.CharBell},
			expected: []byte{},
		},
		{
			name:     "bell and text",
			input:    []byte{readline.CharBell, 'f', 'o', 'o'},
			expected: []byte{readline.CharBell, 'f', 'o', 'o'},
		},
	}

	t.Parallel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdOutBuffer := bytes.NewBuffer([]byte{})

			readline.Stdout = NoopWriteCloser(stdOutBuffer)

			w := &noBellStdout{}
			got, err := w.Write(tc.input)

			assert.NoError(t, err)
			assert.Equal(t, len(tc.expected), got)

			assert.Equal(t, tc.expected, stdOutBuffer.Bytes())

			err = w.Close()
			assert.NoError(t, err)
		})
	}
}
