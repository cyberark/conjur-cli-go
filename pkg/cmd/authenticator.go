package cmd

import (
	"fmt"
	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/spf13/cobra"
	"strings"
)

type authenticatorClient interface {
	EnableAuthenticator(authenticatorType string, serviceID string, enabled bool) error
}

type authenticatorClientFactoryFunc func(*cobra.Command) (authenticatorClient, error)

func authenticatorClientFactory(cmd *cobra.Command) (authenticatorClient, error) {
	return clients.AuthenticatedConjurClientForCommand(cmd)
}

func newEnableCmd(clientFactory authenticatorClientFactoryFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable an authenticator",
		Long: `Enable an authenticator

Examples:

- conjur authenticator enable -i authn-iam/prod
- conjur authenticator enable --id authn-jwt/myVendor`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ID, err := cmd.Flags().GetString("id")
			if err != nil {
				return err
			}

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			authenticatorType, serviceID, err := parseAuthenticatorID(ID)
			if err != nil {
				return err
			}

			err = client.EnableAuthenticator(authenticatorType, serviceID, true)
			if err != nil {
				return err
			}

			cmd.Printf("Authenticator %s is enabled\n", ID)
			return nil
		},
	}

	return cmd
}

func newDisableCmd(clientFactory authenticatorClientFactoryFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable an authenticator",
		Long: `Disable an authenticator

Examples:

- conjur authenticator disable -i authn-iam/prod
- conjur authenticator disable --id authn-jwt/myVendor`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ID, err := cmd.Flags().GetString("id")
			if err != nil {
				return err
			}

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			authenticatorType, serviceID, err := parseAuthenticatorID(ID)
			if err != nil {
				return err
			}

			err = client.EnableAuthenticator(authenticatorType, serviceID, false)
			if err != nil {
				return err
			}

			cmd.Printf("Authenticator %s is disabled\n", ID)
			return nil
		},
	}

	return cmd
}

func parseAuthenticatorID(ID string) (authenticatorType, serviceID string, err error) {
	if ID == "authn-gcp" {
		// For a GCP authenticator, use only authn-gcp; the <service-ID> is not relevant in this use case
		return "gcp", "", nil
	}
	// Split the ID into authenticator type and service ID
	parts := strings.SplitN(ID, "/", 2)
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid authenticator ID format: %s, expected format is 'authenticator_type/service_id'", ID)
	}
	authenticatorType = parts[0]

	// Remove the "authn-" prefix if it exists
	if strings.Contains(authenticatorType, "authn-") {
		authenticatorType = strings.TrimPrefix(authenticatorType, "authn-")
	}

	serviceID = parts[1]
	return authenticatorType, serviceID, nil
}

func newAuthenticatorCommand(clientFactory authenticatorClientFactoryFunc) *cobra.Command {
	authenticatorCmd := &cobra.Command{
		Use:   "authenticator",
		Short: "Manage Secrets Manager authenticators",
	}

	authenticatorCmd.PersistentFlags().StringP("id", "i", "", "(Required) Provide authenticator identifier")
	authenticatorCmd.MarkPersistentFlagRequired("id")

	authenticatorEnable := newEnableCmd(clientFactory)
	authenticatorDisable := newDisableCmd(clientFactory)

	authenticatorCmd.AddCommand(authenticatorEnable)
	authenticatorCmd.AddCommand(authenticatorDisable)
	return authenticatorCmd
}

func init() {
	authenticatorCmd := newAuthenticatorCommand(authenticatorClientFactory)

	rootCmd.AddCommand(authenticatorCmd)
}
