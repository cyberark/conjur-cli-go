package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"

	"github.com/spf13/cobra"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-cli-go/pkg/clients"
)

type createTokenClientFactoryFunc func(*cobra.Command) (createTokenClient, error)
type revokeTokenClientFactoryFunc func(*cobra.Command) (revokeTokenClient, error)
type createHostClientFactoryFunc func(*cobra.Command) (createHostClient, error)

func createTokenClientFactory(cmd *cobra.Command) (createTokenClient, error) {
	return clients.AuthenticatedConjurClientForCommand(cmd)
}
func revokeTokenClientFactory(cmd *cobra.Command) (revokeTokenClient, error) {
	return clients.AuthenticatedConjurClientForCommand(cmd)
}
func createHostClientFactory(cmd *cobra.Command) (createHostClient, error) {
	return clients.AuthenticatedConjurClientForCommand(cmd)
}

type createTokenClient interface {
	CreateToken(durationStr string, hostFactory string, cidr []string, count int) ([]conjurapi.HostFactoryTokenResponse, error)
}
type revokeTokenClient interface {
	DeleteToken(token string) error
}
type createHostClient interface {
	CreateHost(id string, token string) (conjurapi.HostFactoryHostResponse, error)
}

func iPArrayToStingArray(ipArray []net.IP) []string {
	s := make([]string, 0)
	for _, ip := range ipArray {
		s = append(s, ip.String())
	}
	return s
}

func newHostsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hosts",
		Short: "Commands for managing hostfactory hosts",
		Run: func(cmd *cobra.Command, args []string) {
			// Print --help if called without subcommand
			cmd.Help()
		},
	}
}

func newHostsCreateCmd(clientFactory createHostClientFactoryFunc) *cobra.Command {
	hostsCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Use a token to create a host",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := cmd.Flags().GetString("token")
			if err != nil {
				return err
			}
			id, err := cmd.Flags().GetString("id")
			if err != nil {
				return err
			}
			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}
			hostCreateResponse, err := client.CreateHost(id, token)
			if err != nil {
				return err
			}
			indentedResponse, err := json.MarshalIndent(hostCreateResponse, "", "  ")
			if err != nil {
				return err
			}
			cmd.Println(string(indentedResponse))
			return nil
		},
	}

	hostsCreateCmd.Flags().StringP("token", "t", "", "Token")
	hostsCreateCmd.MarkFlagRequired("token")
	hostsCreateCmd.Flags().StringP("id", "i", "", "ID")
	hostsCreateCmd.MarkFlagRequired("id")

	return hostsCreateCmd
}

func newTokensCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tokens",
		Short: "Operations on tokens",
		Run: func(cmd *cobra.Command, args []string) {
			// Print --help if called without subcommand
			cmd.Help()
		},
	}
}

func newTokensCreateCmd(clientFactory createTokenClientFactoryFunc) *cobra.Command {
	tokensCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create one or more tokens",
		Long: `Create one or more host factory tokens. Each token can be used to create
hosts, using hostfactory create hosts.
Valid time units for the --duration flag are "ns", "us" (or "µs"), "ms", "s", "m", "h".

Examples:
- conjur hostfactory tokens create --duration 5m -i cucumber_host_factory_factory
`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			length := len(args)
			if length > 0 {
				// positional args used
			}

			duration, err := cmd.Flags().GetString("duration")
			if err != nil {
				return err
			}
			durationDays, err := cmd.Flags().GetInt64("duration-days")
			if err != nil {
				return err
			}

			durationHours, err := cmd.Flags().GetInt64("duration-hours")
			if err != nil {
				return err
			}

			durationMinutes, err := cmd.Flags().GetInt64("duration-minutes")
			if err != nil {
				return err
			}

			if durationDays < 0 || durationHours < 0 || durationMinutes < 0 {
				return fmt.Errorf("duration values must be positive")
			}

			durationIsSet := cmd.Flags().Changed("duration")
			granularDurationIsSet := cmd.Flags().Changed("duration-days") ||
				cmd.Flags().Changed("duration-hours") ||
				cmd.Flags().Changed("duration-minutes")

			if durationIsSet && granularDurationIsSet {
				return fmt.Errorf("duration can not be used with duration-days, duration-hours or duration-minutes")
			}

			if granularDurationIsSet {
				duration = fmt.Sprintf("%dh%dm", durationDays*24+durationHours, durationMinutes)
			}

			hostfactoryName, err := cmd.Flags().GetString("hostfactory-id")
			if err != nil {
				return err
			}

			// BEGIN COMPATIBILITY WITH PYTHON CLI
			if hostfactoryName == "" {
				hostfactoryName, err = cmd.Flags().GetString("hostfactoryid")
				if err != nil {
					return err
				}
			}

			// Adding this explicit check temporarily because if we
			// mark the non-deprecated flag as required then the cmd
			// will fail when the deprecated flag is used.
			if hostfactoryName == "" {
				return errors.New("Must specify --hostfactory-id")
			}
			// END COMPATIBILITY WITH PYTHON CLI

			cidr, err := cmd.Flags().GetIPSlice("cidr")
			if err != nil {
				return err
			}
			count, err := cmd.Flags().GetInt("count")
			if err != nil {
				return err
			}
			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}
			tokenCreateResponse, err := client.CreateToken(duration, hostfactoryName, iPArrayToStingArray(cidr), count)
			if err != nil {
				return err
			}
			indentedResponse, err := json.MarshalIndent(tokenCreateResponse, "", "  ")
			cmd.Println(string(indentedResponse))
			return err
		},
	}

	tokensCreateCmd.Flags().StringP("duration", "", "10m", "The length of time that the token is valid")

	tokensCreateCmd.Flags().Int64P("duration-days", "", 0, "The component in days of the length of time that the token is valid")
	tokensCreateCmd.Flags().Int64P("duration-hours", "", 0, "The component in hours of the length of time that the token is valid")
	tokensCreateCmd.Flags().Int64P("duration-minutes", "", 0, "The component in minutes of the length of time that the token is valid")
	tokensCreateCmd.Flags().MarkDeprecated("duration-days", "This flag will be removed in a future version.")
	tokensCreateCmd.Flags().MarkDeprecated("duration-hours", "This flag will be removed in a future version.")
	tokensCreateCmd.Flags().MarkDeprecated("duration-minutes", "This flag will be removed in a future version.")
	tokensCreateCmd.Flags().Lookup("duration-days").Hidden = false
	tokensCreateCmd.Flags().Lookup("duration-hours").Hidden = false
	tokensCreateCmd.Flags().Lookup("duration-minutes").Hidden = false

	tokensCreateCmd.Flags().StringP("hostfactory-id", "i", "", "Fully qualified host factory id")

	// BEGIN COMPATIBILITY WITH PYTHON CLI
	// Adds support for 'hostfactoryid' flag to 'hostfactory tokens create' command

	// Uncomment this line when the deprecated flag is removed
	//tokensCreateCmd.MarkFlagRequired("hostfactory-id")

	tokensCreateCmd.Flags().StringP("hostfactoryid", "", "", "")
	tokensCreateCmd.Flags().MarkDeprecated("hostfactoryid", "Use --hostfactory-id instead")
	tokensCreateCmd.Flags().Lookup("hostfactoryid").Hidden = false
	// END COMPATIBILITY WITH PYTHON CLI

	ip, _, _ := net.ParseCIDR("0.0.0.0/0")
	ips := []net.IP{ip}
	tokensCreateCmd.Flags().IPSliceP("cidr", "c", ips, "A comma-delimited list of CIDR addresses to restrict token to")
	tokensCreateCmd.Flags().IntP("count", "n", 1, "Number of tokens to create")
	return tokensCreateCmd
}

func newTokensRevokeCmd(clientFactory revokeTokenClientFactoryFunc) *cobra.Command {
	tokensRevokeCmd := &cobra.Command{
		Use:   "revoke",
		Short: "Revoke (delete) a token",

		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := cmd.Flags().GetString("token")
			if err != nil {
				return err
			}

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}
			err = client.DeleteToken(token)
			if err != nil {
				return err
			}
			return err
		},
	}

	tokensRevokeCmd.Flags().StringP("token", "t", "", "The token to revoke")
	tokensRevokeCmd.MarkFlagRequired("token")

	return tokensRevokeCmd
}

func newHostFactoryCmd(createTokenClientFactory createTokenClientFactoryFunc,
	revokeTokenClientFactory revokeTokenClientFactoryFunc,
	createHostClientFactory createHostClientFactoryFunc,
) *cobra.Command {
	hostfactoryCmd := &cobra.Command{
		Use:   "hostfactory",
		Short: "Manage host factories",
	}
	hostsCmd := newHostsCmd()
	hostfactoryCmd.AddCommand(hostsCmd)
	hostsCreateCmd := newHostsCreateCmd(createHostClientFactory)
	hostsCmd.AddCommand(hostsCreateCmd)

	tokensCmd := newTokensCmd()
	hostfactoryCmd.AddCommand(tokensCmd)

	tokensCreateCmd := newTokensCreateCmd(createTokenClientFactory)
	tokensCmd.AddCommand(tokensCreateCmd)

	tokensRevokeCmd := newTokensRevokeCmd(revokeTokenClientFactory)
	tokensCmd.AddCommand(tokensRevokeCmd)

	// BEGIN COMPATIBILITY WITH PYTHON CLI
	// Adds the 'create host' and 'create token' commands.

	createCmd := newCreateCmd()
	hostfactoryCmd.AddCommand(createCmd)

	createHostCmd := newCreateHostCmd(createHostClientFactory)
	createCmd.AddCommand(createHostCmd)

	createTokenCmd := newCreateTokenCmd(createTokenClientFactory)
	createCmd.AddCommand(createTokenCmd)

	revokeCmd := newRevokeCmd()
	hostfactoryCmd.AddCommand(revokeCmd)

	revokeTokenCmd := newRevokeTokenCmd(revokeTokenClientFactory)
	revokeCmd.AddCommand(revokeTokenCmd)

	// END COMPATIBILITY WITH PYTHON CLI

	return hostfactoryCmd
}

func init() {
	hostfactoryCmd := newHostFactoryCmd(createTokenClientFactory, revokeTokenClientFactory, createHostClientFactory)
	rootCmd.AddCommand(hostfactoryCmd)
}

// BEGIN COMPATIBILITY WITH PYTHON CLI
func newCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "DEPRECATED: Use hostfactory (hosts / tokens) create",
		Run: func(cmd *cobra.Command, args []string) {
			// Print --help if called without subcommand
			cmd.Help()
		},
	}
}

func newCreateHostCmd(clientFactory createHostClientFactoryFunc) *cobra.Command {
	realCmd := newHostsCreateCmd(clientFactory)
	realCmd.Use = "host"
	realCmd.Short = "DEPRECATED: Use hostfactory hosts create"

	return realCmd
}

func newCreateTokenCmd(clientFactory createTokenClientFactoryFunc) *cobra.Command {
	realCmd := newTokensCreateCmd(clientFactory)
	realCmd.Use = "token"
	realCmd.Short = "DEPRECATED: Use hostfactory tokens create"

	return realCmd
}

func newRevokeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "revoke",
		Short: "DEPRECATED: Use hostfactory tokens revoke",
		Run: func(cmd *cobra.Command, args []string) {
			// Print --help if called without subcommand
			cmd.Help()
		},
	}
}

func newRevokeTokenCmd(clientFactory revokeTokenClientFactoryFunc) *cobra.Command {
	realCmd := newTokensRevokeCmd(clientFactory)
	realCmd.Use = "token"
	realCmd.Short = "DEPRECATED: Use hostfactory tokens revoke"

	return realCmd
}

// END COMPATIBILITY WITH PYTHON CLI
