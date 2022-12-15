package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/stretchr/testify/assert"
)

func TestStoreCredentials(t *testing.T) {
	config := setupConfig(t)

	t.Run("Creates file if it does not exist", func(t *testing.T) {
		os.Remove(config.NetRCPath)

		err := StoreCredentials(config, "login", "apiKey")
		assert.NoError(t, err)

		contents, err := os.ReadFile(config.NetRCPath)
		assert.NoError(t, err)
		assert.Contains(t, string(contents), config.ApplianceURL)
		assert.Contains(t, string(contents), "apiKey")
	})

	t.Run("Creates machine if it does not exist", func(t *testing.T) {
		os.Remove(config.NetRCPath)
		_, err := os.Create(config.NetRCPath)
		assert.NoError(t, err)

		err = StoreCredentials(config, "login", "apiKey")
		assert.NoError(t, err)

		contents, err := os.ReadFile(config.NetRCPath)
		assert.NoError(t, err)
		assert.Contains(t, string(contents), config.ApplianceURL)
		assert.Contains(t, string(contents), "apiKey")
	})

	t.Run("Updates machine if it exists", func(t *testing.T) {
		os.Remove(config.NetRCPath)
		initialContent := `
machine http://conjur/authn
	login admin
	password password`

		err := os.WriteFile(config.NetRCPath, []byte(initialContent), 0644)
		assert.NoError(t, err)

		err = StoreCredentials(config, "login", "apiKey")
		assert.NoError(t, err)

		contents, err := os.ReadFile(config.NetRCPath)
		assert.NoError(t, err)
		assert.Contains(t, string(contents), config.ApplianceURL)
		assert.Contains(t, string(contents), "apiKey")
	})
}

func TestPurgeCredentials(t *testing.T) {
	config := setupConfig(t)

	t.Run("Removes machine if it exists", func(t *testing.T) {
		initialContent := `
machine http://conjur/authn
	login admin
	password password`

		err := os.WriteFile(config.NetRCPath, []byte(initialContent), 0644)
		assert.NoError(t, err)

		err = PurgeCredentials(config)
		assert.NoError(t, err)

		contents, err := os.ReadFile(config.NetRCPath)
		assert.NoError(t, err)
		assert.NotContains(t, string(contents), config.ApplianceURL)
	})

	t.Run("Does not error if machine does not exist", func(t *testing.T) {
		os.Remove(config.NetRCPath)
		_, err := os.Create(config.NetRCPath)
		assert.NoError(t, err)

		err = PurgeCredentials(config)
		assert.NoError(t, err)
	})

	t.Run("Does not error if file does not exist", func(t *testing.T) {
		os.Remove(config.NetRCPath)

		err := PurgeCredentials(config)
		assert.NoError(t, err)
	})

	t.Run("Removes oidc token file", func(t *testing.T) {
		_, err := os.Create(config.OidcTokenPath)
		assert.NoError(t, err)

		err = PurgeCredentials(config)
		assert.NoError(t, err)

		_, err = os.Stat(config.OidcTokenPath)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("Remove oidc token file at default location", func(t *testing.T) {
		defaultConfig := conjurapi.Config{
			ApplianceURL: "http://conjur",
		}

		err := os.MkdirAll(os.ExpandEnv("$HOME/.conjur"), 0755)
		assert.NoError(t, err)
		_, err = os.Create(os.ExpandEnv("$HOME/.conjur/oidc_token"))
		assert.NoError(t, err)

		err = PurgeCredentials(defaultConfig)
		assert.NoError(t, err)

		_, err = os.Stat("$HOME/.conjur/oidc_token")
		assert.True(t, os.IsNotExist(err))
	})
}

func setupConfig(t *testing.T) conjurapi.Config {
	tempDir := t.TempDir()
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})
	return conjurapi.Config{
		ApplianceURL:  "http://conjur",
		NetRCPath:     filepath.Join(tempDir, ".netrc"),
		OidcTokenPath: filepath.Join(tempDir, "oidc-token"),
	}
}
