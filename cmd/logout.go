package cmd

import (
	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := conjurapi.LoadConfig()
		if err != nil {
			return err
		}

		err = purgeCredentials(config)
		if err != nil {
			return err
		}

		cmd.Println("Logged out")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
