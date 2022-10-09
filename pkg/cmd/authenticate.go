package cmd

import (
	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/cyberark/conjur-cli-go/pkg/utils"

	"github.com/spf13/cobra"
)

var authenticateCmd = &cobra.Command{
	Use:          "authenticate",
	Short:        "A brief description of your command",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		conjurClient, err := clients.AuthenticatedConjurClientForCommand(cmd)
		if err != nil {
			return err
		}
		// TODO: If there's no credentials then login before storing and authenticating

		//  TODO: Once again I should be able to create an unauthenticated client, and then try to authenticate them using whatever is available!
		data, err := conjurClient.InternalAuthenticate()
		if err != nil {
			return err
		}

		if prettyData, err := utils.PrettyPrintJSON(data); err == nil {
			data = prettyData
		}
		cmd.Println(string(data))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(authenticateCmd)
}
