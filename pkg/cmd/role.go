package cmd

import (
	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/cyberark/conjur-cli-go/pkg/utils"
	"github.com/spf13/cobra"
)

type roleClient interface {
	RoleExists(roleID string) (bool, error)
	Role(roleID string) (role map[string]interface{}, err error)
	RoleMembers(roleID string) (members []map[string]interface{}, err error)
	RoleMembershipsAll(roleID string) (memberships []string, err error)
}

type roleClientFactoryFunc func(*cobra.Command) (roleClient, error)

func roleClientFactory(cmd *cobra.Command) (roleClient, error) {
	return clients.AuthenticatedConjurClientForCommand(cmd)
}

var roleCmd = &cobra.Command{
	Use:   "role",
	Short: "Manage roles",
	Run: func(cmd *cobra.Command, args []string) {
		// Print --help if called without subcommand
		cmd.Help()
	},
}

func newRoleExistsCmd(clientFactory roleClientFactoryFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exists <role-id>",
		Short: "Check if a role exists",
		Long: `Check if a role exists
		
This command requires a [role-id] and includes an optional [--json] 
flag to return a JSON-formatted result.
		
Examples:

- conjur role exists dev:host:somehost
- conjur role exists --json dev:layer:somelayer`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				cmd.Help()
				return nil
			}

			roleID := args[0]

			jsonFlag, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			result, err := client.RoleExists(roleID)
			if err != nil {
				return err
			}

			if jsonFlag {
				data := map[string]interface{}{
					"exists": result,
				}

				prettyData, err := utils.PrettyPrintToJSON(data)
				if err != nil {
					return err
				}

				cmd.Println(prettyData)

			} else {
				cmd.Println(result)
			}

			return nil
		},
	}

	cmd.Flags().Bool("json", false, "Output a JSON response with field 'exists'")

	return cmd
}

func newRoleShowCmd(clientFactory roleClientFactoryFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "show <role-id>",
		Short: "Show a role",
		Long: `Show a role
		
This command requires one argument, a [role-id].

Examples:

-   conjur role show dev:user:someuser
-   conjur role show dev:group:somegroup`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				cmd.Help()
				return nil
			}

			roleID := args[0]

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			result, err := client.Role(roleID)
			if err != nil {
				return err
			}

			prettyResult, err := utils.PrettyPrintToJSON(result)
			if err != nil {
				return err
			}

			cmd.Println(prettyResult)

			return nil
		},
	}
}

func newRoleMembersCmd(clientFactory roleClientFactoryFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "members <role-id>",
		Short: "List members within a role",
		Long: `List members within a role

This command requires a [role-id] and includes an optional [-v|--verbose] 
flag to return the full members object.

Examples:

-   conjur role members dev:user:alice
-   conjur role members --verbose dev:host:bob`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				cmd.Help()
				return nil
			}

			roleID := args[0]

			verboseFlag, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			result, err := client.RoleMembers(roleID)
			if err != nil {
				return err
			}

			var members []interface{}
			if verboseFlag {
				for _, m := range result {
					tmp := make(map[string]interface{}, 0)
					tmp["role"], tmp["member"], tmp["admin_option"] = m["role"], m["member"], m["admin_option"]
					members = append(members, tmp)
				}
			} else {
				for _, element := range result {
					members = append(members, element["member"].(string))
				}
			}

			prettyResult, _ := utils.PrettyPrintToJSON(members)
			if err != nil {
				return err
			}

			cmd.Println(prettyResult)

			return nil
		},
	}

	cmd.Flags().BoolP("verbose", "v", false, "Display verbose members object")

	return cmd
}

func newRoleMembershipsCmd(clientFactory roleClientFactoryFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "memberships <role-id>",
		Short: "List memberships of a role",
		Long: `List memberships of a role
		
This command requires one argument, a [role-id].

Examples:

-   conjur role memberships dev:layer:somelayer
-   conjur role memberships dev:group:somegroup`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				cmd.Help()
				return nil
			}

			roleID := args[0]

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			result, err := client.RoleMembershipsAll(roleID)
			if err != nil {
				return err
			}

			prettyResult, _ := utils.PrettyPrintToJSON(result)
			if err != nil {
				return err
			}

			cmd.Println(prettyResult)

			return nil
		},
	}
}

func init() {
	rootCmd.AddCommand(roleCmd)

	roleExistsCmd := newRoleExistsCmd(roleClientFactory)
	roleShowCmd := newRoleShowCmd(roleClientFactory)
	roleMembersCmd := newRoleMembersCmd(roleClientFactory)
	roleMembershipsCmd := newRoleMembershipsCmd(roleClientFactory)

	roleCmd.AddCommand(roleExistsCmd)
	roleCmd.AddCommand(roleShowCmd)
	roleCmd.AddCommand(roleMembersCmd)
	roleCmd.AddCommand(roleMembershipsCmd)
}
