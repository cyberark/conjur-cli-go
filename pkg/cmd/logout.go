package cmd

import (
	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-cli-go/pkg/output"

	"github.com/spf13/cobra"
)

type configLoaderFn func() (conjurapi.Config, error)

func newLogoutCmd(loadConfig configLoaderFn) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "logout",
		Short:        "Log out the user and delete cached credentials.",
		Long:         `Log out the user and delete the credentials cached in the operating system user's credential storage or .netrc file.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := loadConfig()
			if err != nil {
				return err
			}

			err = conjurapi.PurgeCredentials(config)
			if err != nil {
				return err
			}

			return output.Plain(cmd, map[string]any{"status": "logged_out"}, func() error {
				cmd.Println("Logged out")
				return nil
			})
		},
	}
	return cmd
}

func init() {
	logoutCmd := newLogoutCmd(conjurapi.LoadConfig)
	rootCmd.AddCommand(logoutCmd)
}
