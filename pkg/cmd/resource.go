package cmd

import (
	"encoding/json"

	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/spf13/cobra"
)

type resourceClient interface {
	Resource(resourceID string) (resource map[string]interface{}, err error)
	PermittedRoles(resourceID, privilege string) ([]string, error)
}

type resourceClientFactoryFunc func(*cobra.Command) (resourceClient, error)

func resourceClientFactory(cmd *cobra.Command) (resourceClient, error) {
	return clients.AuthenticatedConjurClientForCommand(cmd)
}

var resourceCmd = &cobra.Command{
	Use:   "resource",
	Short: "Manage resources",
	Run: func(cmd *cobra.Command, args []string) {
		// Print --help if called without subcommand
		cmd.Help()
	},
}

func newResourceExistsCmd(clientFactory resourceClientFactoryFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "exists",
		Short: "Checks if a resource exists",
		Long: `Checks if a resource exists, given a [resource-id].
		
Examples:

- conjur resource exists dev:variable:somevariable
- conjur resource exists dev:host:somehost`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var resourceID string

			if len(args) < 1 {
				cmd.Help()
				return nil
			}

			resourceID = args[0]

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			result, err := client.Resource(resourceID)
			if err != nil {
				return err
			}

			cmd.Println(result != nil)

			return nil
		},
	}
}

func newResourcePermittedRolesCmd(clientFactory resourceClientFactoryFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "permitted-roles",
		Short: "List roles with a specified privilege on the resource",
		Long: `List roles with a specified privilege on the resource
		
This command requires two arguments, a [resource-id] and a [privilege].

Examples:

-   conjur resource permitted-roles dev:variable:somevariable execute
-   conjur resource permitted-roles dev:user:someuser read`,
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

			result, err := client.PermittedRoles(resourceID, privilege)
			if err != nil {
				return err
			}

			prettyResult, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return err
			}

			cmd.Println(string(prettyResult))

			return nil
		},
	}
}

func newResourceShowCmd(clientFactory resourceClientFactoryFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show a resource",
		Long: `Show a resource
		
This command requires one argument, a [resource-id].

Examples:

-   conjur resource show dev:user:someuser
-   conjur resource show dev:host:somehost`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var resourceID string

			if len(args) < 1 {
				cmd.Help()
				return nil
			}

			resourceID = args[0]

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			result, err := client.Resource(resourceID)
			if err != nil {
				return err
			}

			prettyResult, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return err
			}

			cmd.Println(string(prettyResult))

			return nil
		},
	}
}

func init() {
	rootCmd.AddCommand(resourceCmd)

	resourceExistsCmd := newResourceExistsCmd(resourceClientFactory)
	resourcePermittedRolesCmd := newResourcePermittedRolesCmd(resourceClientFactory)
	resourceShowCmd := newResourceShowCmd(resourceClientFactory)

	resourceCmd.AddCommand(resourceExistsCmd)
	resourceCmd.AddCommand(resourcePermittedRolesCmd)
	resourceCmd.AddCommand(resourceShowCmd)
}
