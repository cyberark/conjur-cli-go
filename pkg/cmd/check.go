package cmd

import (
	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/spf13/cobra"
)

type checkClient interface {
	CheckPermission(resourceID string, privilege string) (bool, error)
	CheckPermissionForRole(resourceID string, roleID string, privilege string) (bool, error)
}

type checkClientFactoryFunc func(*cobra.Command) (checkClient, error)

func checkClientFactory(cmd *cobra.Command) (checkClient, error) {
	return clients.AuthenticatedConjurClientForCommand(cmd)
}

func newCheckCmd(clientFactory checkClientFactoryFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check a role's privilege on a resource",
		Long: `Check a role's privilege on a resource

This command requires a [resource-id] and a [privilege].

By default, this command checks if the currently authenticated
user has privilege over the resource. When the optional [-r|--role]
flag is provided, the command checks if the specified role has
privilege over the resource.

Examples:

- conjur check dev:host:somehost write
- conjur check -r user:someuser dev:variable:somevariable read
- conjur check -r dev:user:someuser dev:variable:somevariable read`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var resourceID string
			var privilege string

			if len(args) < 2 {
				cmd.Help()
				return nil
			}

			resourceID, privilege = args[0], args[1]

			roleID, err := cmd.Flags().GetString("role")
			if err != nil {
				return err
			}

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			var result bool

			if len(roleID) == 0 {
				result, err = client.CheckPermission(resourceID, privilege)
			} else {
				result, err = client.CheckPermissionForRole(resourceID, roleID, privilege)
			}
			if err != nil {
				return err
			}

			cmd.Println(result)

			return nil
		},
	}

	cmd.Flags().StringP("role", "r", "", "Partially- or fully-qualified role ID to check privilege for")

	return cmd
}

func init() {
	checkCmd := newCheckCmd(checkClientFactory)

	rootCmd.AddCommand(checkCmd)
}
