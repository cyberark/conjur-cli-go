package cmd

import (
	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/cyberark/conjur-cli-go/pkg/output"
	"github.com/spf13/cobra"
)

type resourceClient interface {
	ResourceExists(resourceID string) (bool, error)
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
	cmd := &cobra.Command{
		Use:   "exists <resource-id>",
		Short: "Check if a resource exists",
		Long: `Check if a resource exists, given a [resource-id].
		
Examples:

- conjur resource exists dev:variable:somevariable
- conjur resource exists dev:host:somehost`,
		RunE: func(cmd *cobra.Command, args []string) error {
			args, err := requirePositionalArgs(cmd, args, "resource-id")
			if err != nil {
				return err
			}

			resourceID := args[0]

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			result, err := client.ResourceExists(resourceID)
			if err != nil {
				return err
			}

			return output.Plain(cmd, map[string]interface{}{"exists": result}, func() error {
				return output.PlainBare(cmd, result)
			})
		},
	}

	return cmd
}

func newResourcePermittedRolesCmd(clientFactory resourceClientFactoryFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "permitted-roles <resource-id> <privilege>",
		Short: "List roles with a specified privilege on the resource",
		Long: `List roles with a specified privilege on the resource
		
This command requires two arguments, a [resource-id] and a [privilege].

Examples:

-   conjur resource permitted-roles dev:variable:somevariable execute
-   conjur resource permitted-roles dev:user:someuser read`,
		RunE: func(cmd *cobra.Command, args []string) error {
			args, err := requirePositionalArgs(cmd, args, "resource-id", "privilege")
			if err != nil {
				return err
			}

			resourceID, privilege := args[0], args[1]

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			result, err := client.PermittedRoles(resourceID, privilege)
			if err != nil {
				return err
			}

			return output.PlainJSONOrUnix(cmd, result)
		},
	}
}

func newResourceShowCmd(clientFactory resourceClientFactoryFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "show <resource-id>",
		Short: "Show a resource",
		Long: `Show a resource
		
This command requires one argument, a [resource-id].

Examples:

-   conjur resource show dev:user:someuser
-   conjur resource show dev:host:somehost`,
		RunE: func(cmd *cobra.Command, args []string) error {
			args, err := requirePositionalArgs(cmd, args, "resource-id")
			if err != nil {
				return err
			}

			resourceID := args[0]

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			result, err := client.Resource(resourceID)
			if err != nil {
				return err
			}

			return output.PlainJSONOrUnix(cmd, result)
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
