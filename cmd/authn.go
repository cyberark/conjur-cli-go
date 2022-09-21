package cmd

import (
	"net/http"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-api-go/conjurapi/authn"
	"github.com/cyberark/conjur-cli-go/pkg/utils"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func newPasswordPrompt() *promptui.Prompt {
	return &promptui.Prompt{
		Label: "Please enter your password (it will not be echoed)",
		Mask:  '*',
	}
}

func newUsernamePrompt() *promptui.Prompt {
	return &promptui.Prompt{
		Label: "Enter your username to log into Conjur",
	}
}

func mightString(v string, e error) string {
	if e != nil {
		return ""
	}

	return v
}

func mightBool(v bool, e error) bool {
	if e != nil {
		return false
	}

	return v
}

func authenticatedConjurClientForCommand(cmd *cobra.Command) (*conjurapi.Client, error) {
	var err error

	setCommandStreamsOnPrompt := func(prompt *promptui.Prompt) *promptui.Prompt {
		prompt.Stdin = utils.NoopReadCloser(cmd.InOrStdin())
		prompt.Stdout = utils.NoopWriteCloser(cmd.OutOrStdout())

		return prompt
	}

	// TODO: This is called multiple time because each operation potentially uses a new HTTP client bound to the
	// temporary Conjur client being created at that point in time. We should really not be creating so many Conjur clients
	decorateConjurClient := func(conjurClient *conjurapi.Client) {}

	if mightBool(cmd.Flags().GetBool("verbose")) {
		decorateConjurClient = func(conjurClient *conjurapi.Client) {
			if conjurClient == nil {
				return
			}
			// TODO: check to see if the transport has already been decorated to make this idempotent

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
	}

	// TODO: extract this common code for gathering configuring into a seperate package
	// Some of the code is in conjur-api-go and needs to be made configurable so that you can pass a custom path to .conjurrc

	config, err := conjurapi.LoadConfig()
	if err != nil {
		return nil, err
	}

	var authenticator conjurapi.Authenticator
	conjurClient, err := conjurapi.NewClientFromEnvironment(config)
	decorateConjurClient(conjurClient)
	if err == nil {
		authenticator = conjurClient.GetAuthenticator()
	} else {
		cmd.Printf("warn: %s\n", err)
	}

	if authenticator == nil {
		username, password, err := requestCredentials(setCommandStreamsOnPrompt, "", "")
		if err != nil {
			return nil, err
		}

		loginPair := authn.LoginPair{Login: username, APIKey: password}

		conjurClient, err = conjurapi.NewClient(config)
		if err != nil {
			return nil, err
		}

		decorateConjurClient(conjurClient)
		data, err := conjurClient.Login(loginPair)
		if err != nil {
			return nil, err
		}

		conjurClient, err = conjurapi.NewClientFromKey(config, authn.LoginPair{Login: username, APIKey: string(data)})
		decorateConjurClient(conjurClient)
		if err != nil {
			return nil, err
		}
	}

	return conjurClient, nil
}

// TODO: whenever this is called we should store to .netrc
func requestCredentials(decoratePrompt decoratePromptFunc, username string, password string) (string, string, error) {
	var err error

	if len(username) == 0 {
		usernamePrompt := decoratePrompt(newUsernamePrompt())
		username, err = runPrompt(usernamePrompt)
		if err != nil {
			return "", "", err
		}
	}

	if len(password) == 0 {
		passwordPrompt := decoratePrompt(newPasswordPrompt())
		password, err = runPrompt(passwordPrompt)
		if err != nil {
			return "", "", err
		}
	}

	return username, password, err
}
