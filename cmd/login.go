package cmd

import (
	"net/http"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-api-go/conjurapi/authn"
	"github.com/cyberark/conjur-cli-go/pkg/utils"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:          "login",
	Short:        "A brief description of your command",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		setCommandStreamsOnPrompt := func(prompt *promptui.Prompt) *promptui.Prompt {
			prompt.Stdin = utils.NoopReadCloser(cmd.InOrStdin())
			prompt.Stdout = utils.NoopWriteCloser(cmd.OutOrStdout())

			return prompt
		}

		// TODO: extract this common code for gathering configuring into a seperate package
		// Some of the code is in conjur-api-go and needs to be made configurable so that you can pass a custom path to .conjurrc
		username := mightString(cmd.Flags().GetString("username"))
		password := mightString(cmd.Flags().GetString("password"))

		config, err := conjurapi.LoadConfig()
		if err != nil {
			return err
		}

		// TODO: I should be able to create a client and unauthenticated client
		conjurClient, err := conjurapi.NewClient(config)
		if err != nil {
			return err
		}

		if mightBool(cmd.Flags().GetBool("verbose")) {
			httpClient := conjurClient.GetHttpClient()
			transport := httpClient.Transport
			if transport == nil {
				transport = http.DefaultTransport
			}
			httpClient.Transport = utils.NewDumpTransport(
				transport,
				func(dump []byte) {
					cmd.Println(string(dump))
					cmd.Println()
				},
			)
		}

		username, password, err = askForCredentials(setCommandStreamsOnPrompt, username, password)
		if err != nil {
			return err
		}

		// TODO: Maybe have a specific struct for logging in?
		loginPair := authn.LoginPair{Login: username, APIKey: password}
		data, err := conjurClient.Login(loginPair)
		if err != nil {
			// TODO: Ruby CLI hides actual error and simply says "Unable to authenticate with Conjur. Please check your credentials."
			return err
		}

		err = storeCredentials(config, username, string(data))
		if err != nil {
			return err
		}

		cmd.Println("Logged in")

		// TODO: should we be storing this in .netrc for future use ?
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringP("username", "u", "", "")
	loginCmd.Flags().StringP("password", "p", "", "")
}
