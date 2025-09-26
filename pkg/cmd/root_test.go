package cmd

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRootCmd(t *testing.T) {
	t.Run("version flag", func(t *testing.T) {
		rootCmd := newRootCommand()
		stdout, _, err := executeCommandForTest(t, rootCmd, "--version")

		assert.NoError(t, err)
		currentYear := time.Now().Format("2006")
		assert.Equal(t, "Secrets Manager CLI version unset-unset\n\nCopyright (c) "+currentYear+" CyberArk Software Ltd. All rights reserved.\n<www.cyberark.com>\n", stdout)
	})
}
