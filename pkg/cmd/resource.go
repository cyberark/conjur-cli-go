package cmd

import (
	"encoding/json"

	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/spf13/cobra"
)

type resourceClient interface {
	Resource(resourceID string) (resource map[string]interface{}, err error)
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

var resourcePermittedRolesCmd = &cobra.Command{
	Use:   "permitted-roles",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("resource permitted-roles called")
	},
}

func init() {
	rootCmd.AddCommand(resourceCmd)

	resourceExistsCmd := newResourceExistsCmd(resourceClientFactory)
	resourceShowCmd := newResourceShowCmd(resourceClientFactory)

	resourceCmd.AddCommand(resourceExistsCmd)
	resourceCmd.AddCommand(resourceShowCmd)
	resourceCmd.AddCommand(resourcePermittedRolesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// resourceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// resourceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
