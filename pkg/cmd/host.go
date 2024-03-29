package cmd

import (
	"github.com/cyberark/conjur-cli-go/pkg/clients"

	"github.com/spf13/cobra"
)

type hostClient interface {
	RotateHostAPIKey(hostID string) ([]byte, error)
	RotateCurrentUserAPIKey() ([]byte, error)
}

type hostClientFactoryFunc func(*cobra.Command) (hostClient, error)

func hostClientFactory(cmd *cobra.Command) (hostClient, error) {
	return clients.AuthenticatedConjurClientForCommand(cmd)
}

func newHostCmd(clientFactory hostClientFactoryFunc) *cobra.Command {
	hostCmd := &cobra.Command{
		Use:   "host",
		Short: "Host commands (rotate-api-key)",
		Run: func(cmd *cobra.Command, args []string) {
			// Print --help if called without subcommand
			cmd.Help()
		},
	}

	hostCmd.AddCommand(newHostRotateAPIKeyCmd(clientFactory))

	return hostCmd
}

func newHostRotateAPIKeyCmd(clientFactory hostClientFactoryFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rotate-api-key",
		Short: "Rotate a host's API key",
		Long: `Rotate the API key of the host specified by the [id] parameter or for the currently logged-in host if no [id] is provided.

Examples:
- conjur host rotate-api-key
- conjur host rotate-api-key --id ci-staging
- conjur host rotate-api-key --id host:ci-staging
- conjur host rotate-api-key --id dev:host:ci-staging`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			hostID, err := cmd.Flags().GetString("id")

			if err != nil {
				return err
			}

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			var newAPIKey []byte
			if hostID == "" {
				newAPIKey, err = client.RotateCurrentUserAPIKey()
			} else {
				newAPIKey, err = client.RotateHostAPIKey(hostID)
			}

			if err != nil {
				return err
			}

			cmd.Println(string(newAPIKey))

			return nil
		},
	}

	cmd.Flags().StringP("id", "i", "", "host whose API key should be rotated (e.g. prod-db)")

	return cmd
}

func init() {
	hostCmd := newHostCmd(hostClientFactory)

	rootCmd.AddCommand(hostCmd)
}
