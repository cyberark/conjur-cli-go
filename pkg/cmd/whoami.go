package cmd

import (
	"encoding/json"

	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/cyberark/conjur-cli-go/pkg/output"

	"github.com/spf13/cobra"
)

type whoamiClient interface {
	WhoAmI() ([]byte, error)
}

func whoamiClientFactory(cmd *cobra.Command) (whoamiClient, error) {
	return clients.AuthenticatedConjurClientForCommand(cmd)
}

type whoamiClientFactoryFunc func(*cobra.Command) (whoamiClient, error)

func newWhoamiCommand(clientFactory whoamiClientFactoryFunc) *cobra.Command {
	return &cobra.Command{
		Use:          "whoami",
		Short:        "Display info about the logged in user",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			raw, err := client.WhoAmI()
			if err != nil {
				return err
			}

			var userData any
			if err := json.Unmarshal(raw, &userData); err != nil {
				return err
			}

			// Default (non-TTY, no --output) resolves to FormatPlain, which
			// output.Render deliberately does not handle generically (each
			// command's plain shape differs). whoami's plain shape is the
			// same pretty-JSON text it always printed.
			return output.PlainJSONOrUnix(cmd, userData)
		},
	}
}

func init() {
	whoamiCmd := newWhoamiCommand(whoamiClientFactory)
	rootCmd.AddCommand(whoamiCmd)
}
