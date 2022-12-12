package cmd

import (
	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-cli-go/pkg/storage"

	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:          "logout",
	Short:        "Logs out a user and deletes cached credentials.",
	Long:         `Logs out a user and deletes the credentials cached in the operating system user's .netrc file.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := conjurapi.LoadConfig()
		if err != nil {
			return err
		}

		err = storage.PurgeCredentials(config)
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
