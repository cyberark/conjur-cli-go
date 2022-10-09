package clients

import (
	"fmt"
	"io"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-api-go/conjurapi/authn"
	"github.com/cyberark/conjur-cli-go/pkg/prompts"
	"github.com/cyberark/conjur-cli-go/pkg/storage"
	"github.com/cyberark/conjur-cli-go/pkg/utils"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

type conjurClient interface {
	Login(authn.LoginPair) ([]byte, error)
	GetConfig() conjurapi.Config
	WhoAmI() ([]byte, error)
	InternalAuthenticate() ([]byte, error)
	LoadPolicy(mode conjurapi.PolicyMode, policyID string, policy io.Reader) (*conjurapi.PolicyResponse, error)
	AddSecret(variableID string, secretValue string) error
	RetrieveSecret(variableID string) ([]byte, error)
}

func AuthenticatedConjurClientForCommand(cmd *cobra.Command) (conjurClient, error) {
	var err error

	setCommandStreamsOnPrompt := func(prompt *promptui.Prompt) *promptui.Prompt {
		prompt.Stdin = utils.NoopReadCloser(cmd.InOrStdin())
		prompt.Stdout = utils.NoopWriteCloser(cmd.OutOrStdout())

		return prompt
	}

	// TODO: This is called multiple time because each operation potentially uses a new HTTP client bound to the
	// temporary Conjur client being created at that point in time. We should really not be creating so many Conjur clients
	decorateConjurClient := func(client *conjurapi.Client) {}

	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return nil, err
	}
	if verbose {
		decorateConjurClient = func(client *conjurapi.Client) {
			VerboseLoggingForClient(cmd, client)
		}
	}

	// TODO: extract this common code for gathering configuring into a seperate package
	// Some of the code is in conjur-api-go and needs to be made configurable so that you can pass a custom path to .conjurrc

	config, err := conjurapi.LoadConfig()
	if err != nil {
		return nil, err
	}

	if config.ApplianceURL == "" {
		return nil, fmt.Errorf("%s", "Missing required configuration for Conjur API URL")
	}

	if config.Account == "" {
		return nil, fmt.Errorf("%s", "Missing required configuation for Conjur account")
	}

	var authenticator conjurapi.Authenticator
	client, err := conjurapi.NewClientFromEnvironment(config)
	decorateConjurClient(client)
	if err == nil {
		authenticator = client.GetAuthenticator()
	} else {
		cmd.Printf("warn: %s\n", err)
	}

	if authenticator == nil {
		client, err = conjurapi.NewClient(config)
		if err != nil {
			return nil, err
		}
		decorateConjurClient(client)

		authenticatePair, err := LoginWithOptionalPrompt(setCommandStreamsOnPrompt, client, "", "")
		if err != nil {
			return nil, err
		}

		client, err = conjurapi.NewClientFromKey(config, *authenticatePair)
		decorateConjurClient(client)
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}

func LoginWithOptionalPrompt(
	decoratePrompt func(prompt *promptui.Prompt) *promptui.Prompt,
	client conjurClient,
	username string,
	password string,
) (*authn.LoginPair, error) {
	username, password, err := prompts.AskForCredentials(decoratePrompt, username, password)
	if err != nil {
		return nil, err
	}

	// TODO: Maybe have a specific struct for logging in?
	loginPair := authn.LoginPair{Login: username, APIKey: password}
	data, err := client.Login(loginPair)
	if err != nil {
		// TODO: Ruby CLI hides actual error and simply says "Unable to authenticate with Conjur. Please check your credentials."
		return nil, err
	}

	authenticatePair := &authn.LoginPair{Login: username, APIKey: string(data)}

	err = storage.StoreCredentials(client.GetConfig(), authenticatePair.Login, authenticatePair.APIKey)
	if err != nil {
		return nil, err
	}

	return authenticatePair, nil
}
