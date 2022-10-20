package clients

import (
	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-api-go/conjurapi/authn"
	"github.com/cyberark/conjur-cli-go/pkg/prompts"
	"github.com/cyberark/conjur-cli-go/pkg/storage"
	"github.com/manifoldco/promptui"
)

func Login(conjurClient *conjurapi.Client, decoratePrompt func(prompt *promptui.Prompt) *promptui.Prompt) (*conjurapi.Client, error) {
	authenticatePair, err := LoginWithPromptFallback(decoratePrompt, conjurClient, "", "")
	if err != nil {
		return nil, err
	}

	return conjurapi.NewClientFromKey(conjurClient.GetConfig(), *authenticatePair)
}

// LoginWithPromptFallback attempts to login to Conjur using the username and password provided.
// If either the username or password is missing then a prompt is presented to interactively request
// the missing information from the user
func LoginWithPromptFallback(
	decoratePrompt func(prompt *promptui.Prompt) *promptui.Prompt,
	client *conjurapi.Client,
	username string,
	password string,
) (*authn.LoginPair, error) {
	username, password, err := prompts.MaybeAskForCredentials(decoratePrompt, username, password)
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
