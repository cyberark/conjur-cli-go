package clients

import (
	"errors"
	"fmt"

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

func OidcLogin(conjurClient *conjurapi.Client) (*conjurapi.Client, error) {
	oidcProvider, err := getOidcProviderInfo(conjurClient, conjurClient.GetConfig().ServiceID)
	if err != nil {
		return nil, err
	}

	code, err := handleOpenIDFlow(oidcProvider.RedirectURI, OpenBrowser)
	if err != nil {
		return nil, err
	}

	//TODO: Will this reassignment cause issues?
	conjurClient, err = conjurapi.NewClientFromOidcCode(conjurClient.GetConfig(), code, oidcProvider.Nonce, oidcProvider.CodeVerifier)

	//TODO: What to store?
	// err = StoreCredentials(config, config.ServiceID, ...)
	// if err != nil {
	// 	return nil, err
	// }

	if err != nil {
		return nil, err
	}

	return conjurClient, nil
}

func getOidcProviderInfo(conjurClient *conjurapi.Client, serviceID string) (*conjurapi.OidcProviderResponse, error) {
	if serviceID == "" {
		return nil, fmt.Errorf("%s", "Missing required configuration service ID for OIDC authenticator")
	}

	data, err := conjurClient.ListOidcProviders()
	if err != nil {
		return nil, err
	}

	var oidcProvider *conjurapi.OidcProviderResponse
	for _, provider := range data {
		if provider.ServiceID == serviceID {
			oidcProvider = &provider
		}
	}
	if oidcProvider == nil {
		return nil, fmt.Errorf("OIDC provider with service ID %s not found", serviceID)
	}

	return oidcProvider, nil
}
