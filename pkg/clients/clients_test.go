package clients

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestLoadAndValidateConjurConfig(t *testing.T) {
	t.Run("Returns a config object", func(t *testing.T) {
		config, _ := LoadAndValidateConjurConfig(0)

		assert.IsType(t, conjurapi.Config{}, config)
	})
}

func TestLoadAndValidateConjurConfigTimeout(t *testing.T) {
	tests := []struct {
		name        string
		timeout     time.Duration
		wantSeconds int
	}{{
		name:        "no timeout flag use env variable",
		timeout:     0,
		wantSeconds: 10,
	}, {
		name:        "timeout flag overrides env variable",
		timeout:     5 * time.Minute,
		wantSeconds: 300,
	}}
	assert.NoError(t, os.Setenv("CONJUR_HTTP_TIMEOUT", "10"))
	assert.NoError(t, os.Setenv("CONJUR_ACCOUNT", "conjur"))
	assert.NoError(t, os.Setenv("CONJUR_APPLIANCE_URL", "https://conjur"))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := LoadAndValidateConjurConfig(tt.timeout)
			assert.NoError(t, err, fmt.Sprintf("LoadAndValidateConjurConfig(%v)", tt.timeout))
			assert.Equal(t, tt.wantSeconds, config.HTTPTimeout, fmt.Sprintf("LoadAndValidateConjurConfig(%v)", tt.timeout))
		})
	}
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

func TestGetTimeout(t *testing.T) {
	five := 5 * time.Second
	tests := []struct {
		name        string
		timeout     *time.Duration
		wantTimeout time.Duration
	}{{
		name:        "no timeout flag",
		timeout:     nil,
		wantTimeout: 0,
	}, {
		name:        "timeout flag",
		timeout:     &five,
		wantTimeout: five,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().Duration("timeout", time.Second, "timeout for the request")
			if tt.timeout != nil {
				assert.NoError(t, cmd.ParseFlags([]string{"--timeout", tt.timeout.String()}))
			}
			gotTimeout, err := GetTimeout(cmd)
			assert.NoError(t, err, fmt.Sprintf("GetTimeout(%v)", tt.wantTimeout))
			assert.Equalf(t, tt.wantTimeout, gotTimeout, "GetTimeout(%v)", tt.wantTimeout)
		})
	}
}
