package cmd

import (
	"errors"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/cyberark/conjur-cli-go/pkg/utils"
	"github.com/spf13/cobra"
)

type listClient interface {
	Resources(filter *conjurapi.ResourceFilter) ([]map[string]interface{}, error)
}

type listClientFactoryFunc func(*cobra.Command) (listClient, error)

func listClientFactory(cmd *cobra.Command) (listClient, error) {
	return clients.AuthenticatedConjurClientForCommand(cmd)
}

func getResourceFilterObject(cmd *cobra.Command) (*conjurapi.ResourceFilter, error) {
	kind, err := cmd.Flags().GetString("kind")
	if err != nil {
		return nil, err
	}
	search, err := cmd.Flags().GetString("search")
	if err != nil {
		return nil, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return nil, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return nil, err
	}
	role, err := cmd.Flags().GetString("role")
	if err != nil {
		return nil, err
	}

	return &conjurapi.ResourceFilter{
		Kind:   kind,
		Search: search,
		Limit:  limit,
		Offset: offset,
		Role: role,
	}, nil
}

func newListCmd(clientFactory listClientFactoryFunc, roleClientFactory roleClientFactoryFunc, resourceClientFactory resourceClientFactoryFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List resources visible to the currently logged-in user",
		Long: `List resources visible to the currently logged-in user.
		
Optional flags can be used to narrow down specific resources.

Examples:
- List all resources : conjur list
- List all users     : conjur list -k user
- List first 5 users : conjur list -k user -l 5
- List next 5 users  : conjur list -k user -l 5 -o 5
- List staging hosts : conjur list -k host -s staging`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			// BEGIN COMPATIBILITY WITH PYTHON CLI
			permittedRoles, err := cmd.Flags().GetString("permitted-roles")
			if err != nil {
				return err
			}

			if permittedRoles != "" {
				privilege, err := cmd.Flags().GetString("privilege")
				if err != nil {
					return err
				}

				if privilege == "" {
					return errors.New("Must specify --privilege when using --permitted-roles")
				}

				realCmd := newResourcePermittedRolesCmd(resourceClientFactory)

				return realCmd.RunE(cmd, []string{permittedRoles, privilege})
			}

			membersOf, err := cmd.Flags().GetString("members-of")
			if err != nil {
				return err
			}

			if membersOf != "" {
				realCmd := newRoleMembersCmd(roleClientFactory)

				return realCmd.RunE(cmd, []string{membersOf})
			}
			// END COMPATIBILITY WITH PYTHON CLI

			rf, err := getResourceFilterObject(cmd)
			if err != nil {
				return err
			}

			resources, err := client.Resources(rf)
			if err != nil {
				return err
			}

			inspect, err := cmd.Flags().GetBool("inspect")
			if err != nil {
				return err
			}

			var prettyResult string

			if inspect {
				prettyResult, err = utils.PrettyPrintToJSON(resources)
			} else {
				resourceIDs := make([]string, 0)

				for _, element := range resources {
					resourceIDs = append(resourceIDs, element["id"].(string))
				}

				prettyResult, err = utils.PrettyPrintToJSON(resourceIDs)
			}

			if err != nil {
				return err
			}

			cmd.Println(prettyResult)

			return nil
		},
	}

	cmd.Flags().StringP("kind", "k", "", "Filter by kind")
	cmd.Flags().StringP("search", "s", "", "Full-text search on resourceID and annotation values")
	cmd.Flags().IntP("limit", "l", 0, "Maximum number of records to return")
	cmd.Flags().IntP("offset", "o", 0, "Offset to start from")
	cmd.Flags().StringP("role", "r", "", "Role whose resource list you want to view")
	cmd.Flags().BoolP("inspect", "i", false, "Show resource details")

	// BEGIN COMPATIBILITY WITH PYTHON CLI
	cmd.Flags().StringP("members-of", "m", "", "List members within a role")
	cmd.Flags().MarkDeprecated("members-of", "Use 'role members' instead")
	cmd.Flags().Lookup("members-of").Hidden = false

	// Must add verbose flag to allow it to be read when calling 'role members'
	cmd.Flags().BoolP("verbose", "v", false, "Display verbose members object")
	cmd.Flags().Lookup("verbose").Hidden = true

	cmd.Flags().StringP("permitted-roles", "", "", "Retrieve roles that have privileges on a resource")
	cmd.Flags().MarkDeprecated("permitted-roles", "Use 'resource permitted-roles' instead")
	cmd.Flags().Lookup("permitted-roles").Hidden = false

	cmd.Flags().StringP("privilege", "p", "", "Use with --permitted-roles to specify the privilege you are querying")
	cmd.Flags().MarkDeprecated("privilege", "Use 'resource permitted-roles' instead")
	cmd.Flags().Lookup("privilege").Hidden = false
	// END COMPATIBILITY WITH PYTHON CLI

	return cmd
}

func init() {
	listCmd := newListCmd(listClientFactory, roleClientFactory, resourceClientFactory)
	rootCmd.AddCommand(listCmd)
}
