package cmd

import (
	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/spf13/cobra"
)

type checkClient interface {
	CheckPermission(resourceID, privilege string) (bool, error)
}

type checkClientFactoryFunc func(*cobra.Command) (checkClient, error)

func checkClientFactory(cmd *cobra.Command) (checkClient, error) {
	return clients.AuthenticatedConjurClientForCommand(cmd)
}

func newCheckCmd(clientFactory checkClientFactoryFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Check for a privilege on a resource",
		Long: `Check whether the currently logged-in user has a given [privilege] on a resource specified by a [resource-id].

Examples:

- conjur check dev:variable:somevariable read
- conjur check dev:host:somehost write`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var resourceID string
			var privilege string

			if len(args) < 2 {
				cmd.Help()
				return nil
			}

			resourceID, privilege = args[0], args[1]

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			result, err := client.CheckPermission(resourceID, privilege)
			if err != nil {
				return err
			}

			cmd.Println(result)

			return nil
		},
	}
}

func init() {
	checkCmd := newCheckCmd(checkClientFactory)
	rootCmd.AddCommand(checkCmd)
}
