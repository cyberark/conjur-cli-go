package cmd

import (
	"github.com/cyberark/conjur-cli-go/pkg/clients"

	"github.com/spf13/cobra"
)

type pubKeysClient interface {
	PublicKeys(kind string, identifier string) ([]byte, error)
}

func pubKeysClientFactory(cmd *cobra.Command) (pubKeysClient, error) {
	return clients.AuthenticatedConjurClientForCommand(cmd)
}

type pubKeysClientFactoryFunc func(*cobra.Command) (pubKeysClient, error)

func newPubKeysCommand(clientFactory pubKeysClientFactoryFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pubkeys <username>",
		Short: "Display the public keys associated with a user",
		Long: `Display the public keys for a given [username].

Examples:
- conjur pubkeys alice`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				cmd.Help()
				return nil
			}

			username := args[0]

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			pubKeysData, err := client.PublicKeys("user", username)
			if err != nil {
				return err
			}

			cmd.Println(string(pubKeysData))

			return nil
		},
	}

	return cmd
}

func init() {
	config := clients.LoadConfigOrDefault()
	if config.IsSelfHosted() || config.IsConjurOSS() {
		pubKeysCmd := newPubKeysCommand(pubKeysClientFactory)
		rootCmd.AddCommand(pubKeysCmd)
	}
}
