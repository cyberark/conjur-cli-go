package cmd

import (
	"github.com/cyberark/conjur-cli-go/pkg/utils"
	"github.com/spf13/cobra"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		conjurClient, err := authenticatedConjurClientForCommand(cmd)
		if err != nil {
			return err
		}
		// TODO: If there's no credentials then login before storing and authenticating

		//  TODO: Once again I should be able to create an unauthenticated client, and then try to authenticate them using whatever is available!
		data, err := conjurClient.WhoAmI()
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
	rootCmd.AddCommand(whoamiCmd)
}
