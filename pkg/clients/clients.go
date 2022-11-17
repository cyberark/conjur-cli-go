package clients

import (
	"errors"
	"fmt"
	"io"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-api-go/conjurapi/authn"
	"github.com/cyberark/conjur-cli-go/pkg/prompts"
	"github.com/cyberark/conjur-cli-go/pkg/storage"

	"github.com/spf13/cobra"
)

// ConjurClient is an interface that represents a Conjur client
type ConjurClient interface {
	Login(login string, password string) ([]byte, error)
	GetConfig() conjurapi.Config
	WhoAmI() ([]byte, error)
	RotateHostAPIKey(hostID string) ([]byte, error)
	InternalAuthenticate() ([]byte, error)
	LoadPolicy(mode conjurapi.PolicyMode, policyID string, policy io.Reader) (*conjurapi.PolicyResponse, error)
	AddSecret(variableID string, secretValue string) error
	RetrieveSecret(variableID string) ([]byte, error)
	CheckPermission(resourceID, privilege string) (bool, error)
	ResourceIDs(filter *conjurapi.ResourceFilter) ([]string, error)
}

// LoadAndValidateConjurConfig loads and validate Conjur configuration
func LoadAndValidateConjurConfig() (conjurapi.Config, error) {
	// TODO: extract this common code for gathering configuring into a seperate package
	// Some of the code is in conjur-api-go and needs to be made configurable so that you can pass a custom path to .conjurrc

	config, err := conjurapi.LoadConfig()
	if err != nil {
		return config, err
	}

	if config.ApplianceURL == "" {
		return config, fmt.Errorf("%s", "Missing required configuration for Conjur API URL")
	}

	if config.Account == "" {
		return config, fmt.Errorf("%s", "Missing required configuation for Conjur account")
	}

	return config, err
}

// AuthenticatedConjurClientForCommand attempts to get an authenticated Conjur client by iterating through
// configuration, environment variables and then ultimately falling back on prompting the user for credentials.
func AuthenticatedConjurClientForCommand(cmd *cobra.Command) (ConjurClient, error) {
	var err error

	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return nil, err
	}

	// TODO: This is called multiple time because each operation potentially uses a new HTTP client bound to the
	// temporary Conjur client being created at that point in time. We should really not be creating so many Conjur clients
	// we should just have one then the rest is an attempt to get an authenticator
	decorateConjurClient := func(client *conjurapi.Client) {
		MaybeVerboseLoggingForClient(verbose, cmd, client)
	}

	config, err := LoadAndValidateConjurConfig()
	if err != nil {
		return nil, err
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

		authenticatePair, err := LoginWithPromptFallback(prompts.PromptDecoratorForCommand(cmd), client, "", "")
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

// LoginWithPromptFallback attempts to login to Conjur using the username and password provided.
// If either the username or password is missing then a prompt is presented to interactively request
// the missing information from the user
func LoginWithPromptFallback(
	decoratePrompt prompts.DecoratePromptFunc,
	client ConjurClient,
	username string,
	password string,
) (*authn.LoginPair, error) {
	username, password, err := prompts.MaybeAskForCredentials(decoratePrompt, username, password)
	if err != nil {
		return nil, err
	}

	data, err := client.Login(username, password)
	if err != nil {
		return nil, errors.New("Unable to authenticate with Conjur. Please check your credentials.")
	}

	authenticatePair := &authn.LoginPair{Login: username, APIKey: string(data)}

	err = storage.StoreCredentials(client.GetConfig(), authenticatePair.Login, authenticatePair.APIKey)
	if err != nil {
		return nil, err
	}

	return authenticatePair, nil
}
