package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/cyberark/conjur-cli-go/pkg/clients"

	"github.com/spf13/cobra"
)

type userClient interface {
	RotateUserAPIKey(userID string) ([]byte, error)
	WhoAmI() ([]byte, error)
}

type userClientFactoryFunc func(*cobra.Command) (userClient, error)

func userClientFactory(cmd *cobra.Command) (userClient, error) {
	return clients.AuthenticatedConjurClientForCommand(cmd)
}

func newUserCmd(clientFactory userClientFactoryFunc) *cobra.Command {

	userCmd := &cobra.Command{
		Use:   "user",
		Short: "User commands (change-password, rotate-api-key)",
		Run: func(cmd *cobra.Command, args []string) {
			// Print --help if called without subcommand
			cmd.Help()
		},
	}

	userCmd.AddCommand(newUserChangePasswordCmd(clientFactory))
	userCmd.AddCommand(newUserRotateAPIKeyCmd(clientFactory))

	return userCmd
}

func newUserChangePasswordCmd(clientFactory userClientFactoryFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "change-password",
		Short: "A brief description of your command",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("user change-password called")
		},
	}

	return cmd
}

func newUserRotateAPIKeyCmd(clientFactory userClientFactoryFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rotate-api-key",
		Short: "Rotate a user's API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := cmd.Flags().GetString("user-id")
			if err != nil {
				return err
			}

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			// Retrieve username for logged in user
			if userID == "" {
				userData, err := client.WhoAmI()
				if err != nil {
					return err
				}

				var whoamiResponse whoamiResponse
				err = json.Unmarshal(userData, &whoamiResponse)
				if err != nil {
					return err
				}

				userID = whoamiResponse.Username
			}

			newAPIKey, err := client.RotateUserAPIKey(userID)
			if err != nil {
				return err
			}

			cmd.Println(string(newAPIKey))

			return nil
		},
	}

	cmd.Flags().StringP(
		"user-id", "u", "",
		"ID of user whose API key should be rotated (e.g. alice). Defaults to logged-in user.",
	)

	return cmd
}

func init() {
	userCmd := newUserCmd(userClientFactory)

	rootCmd.AddCommand(userCmd)
}
