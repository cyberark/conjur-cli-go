package cmd

import (
	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/cyberark/conjur-cli-go/pkg/utils"

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

			userData, err := client.WhoAmI()
			if err != nil {
				return err
			}

			userData, err = utils.PrettyPrintJSON(userData)
			if err != nil {
				return err
			}

			cmd.Println(string(userData))

			return nil
		},
	}
}

func init() {
	whoamiCmd := newWhoamiCommand(whoamiClientFactory)
	rootCmd.AddCommand(whoamiCmd)
}
