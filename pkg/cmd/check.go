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
		Long: `Check for a privilege on a resource
		
This command requires two arguments, a [resourceId] and a [privilege].

Examples:

-   conjur check dev:variable:somevariable read
-   conjur check dev:host:somehost write`,
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
