package cmd

import (
	"encoding/json"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/spf13/cobra"
)

type listClient interface {
	ResourceIDs(filter *conjurapi.ResourceFilter) ([]string, error)
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

	return &conjurapi.ResourceFilter{
		Kind:   kind,
		Search: search,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func newListCmd(clientFactory listClientFactoryFunc) *cobra.Command {
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

			rf, err := getResourceFilterObject(cmd)
			if err != nil {
				return err
			}

			data, err := client.ResourceIDs(rf)
			if err != nil {
				return err
			}

			res, err := json.MarshalIndent(data, "", "  ")
			if err != nil {
				return err
			}

			cmd.Println(string(res))

			return nil
		},
	}

	cmd.Flags().StringP("kind", "k", "", "Filter by kind")
	cmd.Flags().StringP("search", "s", "", "Full-text search on resourceID and annotation values")
	cmd.Flags().IntP("limit", "l", 0, "Maximum number of records to return")
	cmd.Flags().IntP("offset", "o", 0, "Offset to start from")

	return cmd
}

func init() {
	listCmd := newListCmd(listClientFactory)
	rootCmd.AddCommand(listCmd)
}
