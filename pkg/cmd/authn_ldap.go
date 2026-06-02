package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/spf13/cobra"
)

type ldapGroupMappingClient interface {
	LdapCreateGroupMapping(serviceID, groupName string, roles []string) (*conjurapi.LdapGroupMappingResponse, error)
	LdapShowGroupMapping(serviceID, groupName string) (*conjurapi.LdapGroupMappingResponse, error)
	LdapListGroupMappings(serviceID string) (*conjurapi.LdapMappingListResponse, error)
	LdapDeleteGroupMapping(serviceID, groupName string) error
}

type ldapUserMappingClient interface {
	LdapCreateUserMapping(serviceID, username string, roles []string) (*conjurapi.LdapUserMappingResponse, error)
	LdapShowUserMapping(serviceID, username string) (*conjurapi.LdapUserMappingResponse, error)
	LdapListUserMappings(serviceID string) (*conjurapi.LdapMappingListResponse, error)
	LdapDeleteUserMapping(serviceID, username string) error
}

type ldapGroupMappingClientFactory func(*cobra.Command) (ldapGroupMappingClient, error)
type ldapUserMappingClientFactory func(*cobra.Command) (ldapUserMappingClient, error)

const (
	ldapOutputTable = "table"
	ldapOutputJSON  = "json"
)

func ldapGroupClientFactory(cmd *cobra.Command) (ldapGroupMappingClient, error) {
	return clients.AuthenticatedConjurClientForCommand(cmd)
}

func ldapUserClientFactory(cmd *cobra.Command) (ldapUserMappingClient, error) {
	return clients.AuthenticatedConjurClientForCommand(cmd)
}

func newLdapGroupsCreateCmd(clientFactory ldapGroupMappingClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <group-name>",
		Short: "Create an LDAP group to Conjur role mapping",
		Long: `Create a mapping between an LDAP group and one or more Conjur roles.
When a user authenticates via LDAP and belongs to this group, they will
automatically receive the mapped Conjur roles.

Examples:
- conjur authn-ldap groups create admins --roles admin-role --id ad-prod
- conjur authn-ldap groups create dev-team --roles dev-role,readonly --id ad-prod`,
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupName := args[0]

			serviceID, err := cmd.Flags().GetString("id")
			if err != nil {
				return err
			}

			rolesStr, err := cmd.Flags().GetString("roles")
			if err != nil {
				return err
			}
			roles := splitRoles(rolesStr)

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			result, err := client.LdapCreateGroupMapping(serviceID, groupName, roles)
			if err != nil {
				return fmt.Errorf("failed to create group mapping '%s': %w", groupName, err)
			}

			return printLdapResult(cmd, result)
		},
	}

	cmd.Flags().StringP("roles", "r", "", "(Required) Comma-separated list of Conjur roles to assign")
	cmd.MarkFlagRequired("roles")

	return cmd
}

func newLdapGroupsShowCmd(clientFactory ldapGroupMappingClientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "show <group-name>",
		Short: "Show an LDAP group mapping",
		Long: `Display the details of an LDAP group mapping, including all mapped Conjur roles.

Examples:
- conjur authn-ldap groups show admins --id ad-prod`,
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupName := args[0]

			serviceID, err := cmd.Flags().GetString("id")
			if err != nil {
				return err
			}

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			result, err := client.LdapShowGroupMapping(serviceID, groupName)
			if err != nil {
				return fmt.Errorf("failed to show group mapping '%s': %w", groupName, err)
			}

			return printLdapResult(cmd, result)
		},
	}
}

func newLdapGroupsListCmd(clientFactory ldapGroupMappingClientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all LDAP group mappings",
		Long: `List all configured LDAP group mappings for a service instance.

Examples:
- conjur authn-ldap groups list --id ad-prod`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceID, err := cmd.Flags().GetString("id")
			if err != nil {
				return err
			}

			output, err := getLdapOutputFormat(cmd)
			if err != nil {
				return err
			}

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			result, err := client.LdapListGroupMappings(serviceID)
			if err != nil {
				return fmt.Errorf("failed to list group mappings: %w", err)
			}

			if output == ldapOutputJSON {
				return printJSON(cmd, result)
			}

			for _, key := range result.Keys {
				cmd.Println(key)
			}
			return nil
		},
	}
}

func newLdapGroupsDeleteCmd(clientFactory ldapGroupMappingClientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <group-name>",
		Short: "Delete an LDAP group mapping",
		Long: `Delete an LDAP group mapping. Users will no longer receive the associated
Conjur roles from this LDAP group at login time.

Examples:
- conjur authn-ldap groups delete admins --id ad-prod`,
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupName := args[0]

			serviceID, err := cmd.Flags().GetString("id")
			if err != nil {
				return err
			}

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			err = client.LdapDeleteGroupMapping(serviceID, groupName)
			if err != nil {
				return fmt.Errorf("failed to delete group mapping '%s': %w", groupName, err)
			}

			cmd.Printf("Success! LDAP group mapping '%s' deleted\n", groupName)
			return nil
		},
	}
}

// --- User Commands ---

func newLdapUsersCreateCmd(clientFactory ldapUserMappingClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <username>",
		Short: "Create an LDAP user to Conjur role mapping",
		Long: `Create a mapping between an individual LDAP user and one or more Conjur roles.
This is used for exceptions that cannot be handled by group-based mappings.

Examples:
- conjur authn-ldap users create jan.k --roles special-role --id ad-prod
- conjur authn-ldap users create admin.user --roles admin-role,audit-role --id ad-prod`,
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			username := args[0]

			serviceID, err := cmd.Flags().GetString("id")
			if err != nil {
				return err
			}

			rolesStr, err := cmd.Flags().GetString("roles")
			if err != nil {
				return err
			}
			roles := splitRoles(rolesStr)

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			result, err := client.LdapCreateUserMapping(serviceID, username, roles)
			if err != nil {
				return fmt.Errorf("failed to create user mapping '%s': %w", username, err)
			}

			return printLdapResult(cmd, result)
		},
	}

	cmd.Flags().StringP("roles", "r", "", "(Required) Comma-separated list of Conjur roles to assign")
	cmd.MarkFlagRequired("roles")

	return cmd
}

func newLdapUsersShowCmd(clientFactory ldapUserMappingClientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "show <username>",
		Short: "Show an LDAP user mapping",
		Long: `Display the details of an LDAP user mapping, including all mapped Conjur roles.

Examples:
- conjur authn-ldap users show jan.k --id ad-prod`,
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			username := args[0]

			serviceID, err := cmd.Flags().GetString("id")
			if err != nil {
				return err
			}

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			result, err := client.LdapShowUserMapping(serviceID, username)
			if err != nil {
				return fmt.Errorf("failed to show user mapping '%s': %w", username, err)
			}

			return printLdapResult(cmd, result)
		},
	}
}

func newLdapUsersListCmd(clientFactory ldapUserMappingClientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all LDAP user mappings",
		Long: `List all configured LDAP user mappings for a service instance.

Examples:
- conjur authn-ldap users list --id ad-prod`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceID, err := cmd.Flags().GetString("id")
			if err != nil {
				return err
			}

			output, err := getLdapOutputFormat(cmd)
			if err != nil {
				return err
			}

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			result, err := client.LdapListUserMappings(serviceID)
			if err != nil {
				return fmt.Errorf("failed to list user mappings: %w", err)
			}

			if output == ldapOutputJSON {
				return printJSON(cmd, result)
			}

			for _, key := range result.Keys {
				cmd.Println(key)
			}
			return nil
		},
	}
}

func newLdapUsersDeleteCmd(clientFactory ldapUserMappingClientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <username>",
		Short: "Delete an LDAP user mapping",
		Long: `Delete an LDAP user mapping. The user will no longer receive the associated
Conjur roles from this individual mapping at login time.

Examples:
- conjur authn-ldap users delete jan.k --id ad-prod`,
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			username := args[0]

			serviceID, err := cmd.Flags().GetString("id")
			if err != nil {
				return err
			}

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			err = client.LdapDeleteUserMapping(serviceID, username)
			if err != nil {
				return fmt.Errorf("failed to delete user mapping '%s': %w", username, err)
			}

			cmd.Printf("Success! LDAP user mapping '%s' deleted\n", username)
			return nil
		},
	}
}

func newLdapGroupsCommand(clientFactory ldapGroupMappingClientFactory) *cobra.Command {
	groupsCmd := &cobra.Command{
		Use:   "groups",
		Short: "Manage LDAP group to Conjur role mappings",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	groupsCmd.AddCommand(newLdapGroupsCreateCmd(clientFactory))
	groupsCmd.AddCommand(newLdapGroupsShowCmd(clientFactory))
	groupsCmd.AddCommand(newLdapGroupsListCmd(clientFactory))
	groupsCmd.AddCommand(newLdapGroupsDeleteCmd(clientFactory))

	return groupsCmd
}

func newLdapUsersCommand(clientFactory ldapUserMappingClientFactory) *cobra.Command {
	usersCmd := &cobra.Command{
		Use:   "users",
		Short: "Manage LDAP user to Conjur role mappings",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	usersCmd.AddCommand(newLdapUsersCreateCmd(clientFactory))
	usersCmd.AddCommand(newLdapUsersShowCmd(clientFactory))
	usersCmd.AddCommand(newLdapUsersListCmd(clientFactory))
	usersCmd.AddCommand(newLdapUsersDeleteCmd(clientFactory))

	return usersCmd
}

func newAuthnLdapCommand(groupFactory ldapGroupMappingClientFactory, userFactory ldapUserMappingClientFactory) *cobra.Command {
	authnLdapCmd := &cobra.Command{
		Use:   "authn-ldap",
		Short: "Manage LDAP authentication group and user mappings",
		Long: `Manage LDAP group-to-role and user-to-role mappings for authn-ldap
service instances.`,
	}

	authnLdapCmd.PersistentFlags().StringP("id", "i", "", "(Required) LDAP authenticator service ID (e.g. ad-prod)")
	authnLdapCmd.MarkPersistentFlagRequired("id")
	authnLdapCmd.PersistentFlags().StringP("output", "o", ldapOutputTable, "Output format (table or json)")

	authnLdapCmd.AddCommand(newLdapGroupsCommand(groupFactory))
	authnLdapCmd.AddCommand(newLdapUsersCommand(userFactory))

	return authnLdapCmd
}

func init() {
	config := clients.LoadConfigOrDefault()
	authnLdapCmd := newAuthnLdapCommand(ldapGroupClientFactory, ldapUserClientFactory)
	if config.IsSelfHosted() || config.IsConjurOSS() {
		rootCmd.AddCommand(authnLdapCmd)
	}
}

func splitRoles(rolesStr string) []string {
	var roles []string
	for _, r := range strings.Split(rolesStr, ",") {
		r = strings.TrimSpace(r)
		if r != "" {
			roles = append(roles, r)
		}
	}
	return roles
}

func printJSON(cmd *cobra.Command, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format response: %w", err)
	}
	cmd.Println(string(data))
	return nil
}

func getLdapOutputFormat(cmd *cobra.Command) (string, error) {
	output, err := cmd.Flags().GetString("output")
	if err != nil {
		return "", err
	}

	switch output {
	case ldapOutputTable, ldapOutputJSON:
		return output, nil
	default:
		return "", fmt.Errorf("invalid output format %q: supported values are %q and %q", output, ldapOutputTable, ldapOutputJSON)
	}
}

func printLdapResult(cmd *cobra.Command, v interface{}) error {
	output, err := getLdapOutputFormat(cmd)
	if err != nil {
		return err
	}

	if output == ldapOutputJSON {
		return printJSON(cmd, v)
	}

	return printLdapTable(cmd, v)
}

func printLdapTable(cmd *cobra.Command, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to format response: %w", err)
	}

	var values map[string]interface{}
	if err := json.Unmarshal(data, &values); err != nil {
		return fmt.Errorf("failed to format response: %w", err)
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "Key\tValue")
	fmt.Fprintln(tw, "---\t-----")

	for _, key := range keys {
		fmt.Fprintf(tw, "%s\t%s\n", key, formatLdapTableValue(values[key]))
	}

	return tw.Flush()
}

func formatLdapTableValue(v interface{}) string {
	switch value := v.(type) {
	case []interface{}:
		items := make([]string, 0, len(value))
		for _, item := range value {
			items = append(items, fmt.Sprint(item))
		}
		return fmt.Sprintf("[%s]", strings.Join(items, ", "))
	case map[string]interface{}:
		bytes, err := json.Marshal(value)
		if err != nil {
			return fmt.Sprint(value)
		}
		return string(bytes)
	default:
		return fmt.Sprint(value)
	}
}
