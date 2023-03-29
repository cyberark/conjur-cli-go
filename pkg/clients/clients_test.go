package clients

import (
	"testing"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestLoadAndValidateConjurConfig(t *testing.T) {
	t.Run("Returns a config object", func(t *testing.T) {
		config, _ := LoadAndValidateConjurConfig()

		assert.IsType(t, conjurapi.Config{}, config)
	})
}

func TestAuthenticatedConjurClientForCommand(t *testing.T) {
	t.Run("Errors when no valid config", func(t *testing.T) {
		var cmd = &cobra.Command{}
		cmd.Flags().Bool("debug", false, "Debug logging enabled")

		client, err := AuthenticatedConjurClientForCommand(cmd)

		// Note: This fails when run inside the dev container because it has a valid config
		assert.Nil(t, client)
		assert.Error(t, err)
	})
}
