package cmd

import (
	"github.com/cyberark/conjur-cli-go/pkg/clients"

	"github.com/spf13/cobra"
)

var hostCmd = &cobra.Command{
	Use:   "host",
	Short: "Host commands (rotate-api-key)",
	Run: func(cmd *cobra.Command, args []string) {
		// Print --help if called without subcommand
		cmd.Help()
	},
}

type hostRotateAPIKeyClient interface {
	RotateHostAPIKey(roleID string) ([]byte, error)
}

func hostRotateAPIKeyClientFactory(cmd *cobra.Command) (hostRotateAPIKeyClient, error) {
	return clients.AuthenticatedConjurClientForCommand(cmd)
}

// NewHostRotateAPIKeyCommand creates a Command instance with injected dependencies.
func NewHostRotateAPIKeyCommand(clientFactory func(*cobra.Command) (hostRotateAPIKeyClient, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "rotate-api-key",
		Short:        "Rotate a host's API key",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			hostID, err := cmd.Flags().GetString("host")
			if err != nil {
				return err
			}

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			newAPIKey, err := client.RotateHostAPIKey(hostID)
			if err != nil {
				return err
			}

			cmd.Println(string(newAPIKey))

			return nil
		},
	}

	cmd.Flags().StringP("host", "", "", "")

	return cmd
}

func init() {
	rootCmd.AddCommand(hostCmd)

	hostRotateAPIKeyCmd := NewHostRotateAPIKeyCommand(hostRotateAPIKeyClientFactory)
	hostCmd.AddCommand(hostRotateAPIKeyCmd)
}
