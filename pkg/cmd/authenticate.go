package cmd

import (
	"encoding/base64"

	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/cyberark/conjur-cli-go/pkg/utils"

	"github.com/spf13/cobra"
)

type authenticateClient interface {
	InternalAuthenticate() ([]byte, error)
}

func authenticateClientFactory(cmd *cobra.Command) (authenticateClient, error) {
	return clients.AuthenticatedConjurClientForCommand(cmd)
}

type authenticateClientFactoryFunc func(*cobra.Command) (authenticateClient, error)

// NewAuthenticateCommand creates an authenticate Command instance with injected dependencies.
func newAuthenticateCommand(clientFactory authenticateClientFactoryFunc) *cobra.Command {
	authenticateCmd := &cobra.Command{
		Use:          "authenticate",
		Short:        "Obtain an access token for the currently logged-in user",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			conjurClient, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			//  TODO: I should be able to create an unauthenticated client, and then try to authenticate them using whatever is available!
			data, err := conjurClient.InternalAuthenticate()
			if err != nil {
				return err
			}

			formatAsHeader, err := cmd.Flags().GetBool("header")
			if err != nil {
				return err
			}

			if formatAsHeader {
				// Base64 encode the result and format as an HTTP Authorization header for the -H option
				cmd.Println("Authorization: Token token=\"" + base64.StdEncoding.EncodeToString(data) + "\"")
				return nil
			}

			prettyData, err := utils.PrettyPrintJSON(data)
			if err != nil {
				return err
			}

			cmd.Println(string(prettyData))
			return nil
		},
	}

	authenticateCmd.Flags().BoolP(
		"header",
		"H",
		false,
		"Base64 encode the result and format as an HTTP Authorization header",
	)

	return authenticateCmd
}

func init() {
	authenticateCmd := newAuthenticateCommand(authenticateClientFactory)
	rootCmd.AddCommand(authenticateCmd)
}
