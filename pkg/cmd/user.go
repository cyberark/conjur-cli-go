package cmd

import (
	"encoding/json"

	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/cyberark/conjur-cli-go/pkg/prompts"

	"github.com/spf13/cobra"
)

type userClient interface {
	RotateUserAPIKey(userID string) ([]byte, error)
	WhoAmI() ([]byte, error)
	ChangeCurrentUserPassword(newPassword string) ([]byte, error)
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

func newUserRotateAPIKeyCmd(clientFactory userClientFactoryFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rotate-api-key",
		Short: "Rotate a user's API key",
		Long: `Rotate the API key of the user specified by the [id] parameter or for the currently logged-in user if no [id] is provided.

Examples:
- conjur user rotate-api-key --id alice
- conjur user rotate-api-key`,
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := cmd.Flags().GetString("id")

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

	cmd.Flags().StringP("id", "i", "", "user whose API key should be rotated (default: logged-in user)")

	return cmd
}

func newUserChangePasswordCmd(clientFactory userClientFactoryFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "change-password",
		Short: "Update the password for the currently logged-in user.",
		Long: `Update the password for the currently logged-in user to [password].

If no password flag is provided, the user will be prompted.

Examples:
- conjur user change-password -p SUp3r$3cr3t!!`,
		RunE: func(cmd *cobra.Command, args []string) error {
			newPassword, err := cmd.Flags().GetString("password")

			setCommandStreamsOnPrompt := prompts.PromptDecoratorForCommand(cmd)

			newPassword, err = prompts.MaybeAskForChangePassword(
				setCommandStreamsOnPrompt,
				newPassword,
			)

			if err != nil {
				return err
			}

			client, err := clientFactory(cmd)
			if err != nil {
				return err
			}

			_, err = client.ChangeCurrentUserPassword(newPassword)
			if err != nil {
				return err
			}

			cmd.Println("Password changed")

			return nil
		},
	}

	cmd.Flags().StringP("password", "p", "", "The new password")

	return cmd
}

func init() {
	userCmd := newUserCmd(userClientFactory)

	rootCmd.AddCommand(userCmd)
}
