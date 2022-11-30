package cmd

import (
	"encoding/base64"
	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/cyberark/conjur-cli-go/pkg/utils"

	"github.com/spf13/cobra"
)

var authenticateCmd = &cobra.Command{
	Use:          "authenticate",
	Short:        "Obtains an authentication token using the current logged-in user",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		conjurClient, err := clients.AuthenticatedConjurClientForCommand(cmd)
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
		} else {
			if prettyData, err := utils.PrettyPrintJSON(data); err == nil {
				data = prettyData
			}
			cmd.Println(string(data))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(authenticateCmd)
	authenticateCmd.Flags().BoolP("header", "H", false,
		"Base64 encode the result and format as an HTTP Authorization header")
}
