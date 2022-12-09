package clients

import (
	"errors"
	"fmt"
	"log"

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

func OidcLogin(conjurClient *conjurapi.Client, username string, password string) (*conjurapi.Client, error) {
	config := conjurClient.GetConfig()

	oidcProvider, err := getOidcProviderInfo(conjurClient, config.ServiceID)
	if err != nil {
		return nil, err
	}

	// If a username and password is provided, attempt to use them to fetch an OIDC code instead of opening a browser
	oidcPromptHandler := OpenBrowser
	if username != "" && password != "" {
		log.Print("Attempting to use username and password to fetch OIDC code. " +
			"This is meant for testing purposes only and should not be used in production.")
		oidcPromptHandler = FetchOidcCodeFromProvider(username, password)
	}

	code, err := handleOpenIDFlow(oidcProvider.RedirectURI, oidcPromptHandler)
	if err != nil {
		return nil, err
	}

	conjurClient, err = conjurapi.NewClientFromOidcCode(config, code, oidcProvider.Nonce, oidcProvider.CodeVerifier)
	if err != nil {
		return nil, err
	}

	// Refreshes the access token and caches it locally
	err = conjurClient.RefreshToken()
	if err != nil {
		return nil, err
	}

	return conjurClient, nil
}

func getOidcProviderInfo(conjurClient ConjurClient, serviceID string) (*conjurapi.OidcProviderResponse, error) {
	if serviceID == "" {
		return nil, fmt.Errorf("%s", "Missing required configuration service ID for OIDC authenticator")
	}

	data, err := conjurClient.ListOidcProviders()
	if err != nil {
		return nil, err
	}

	// Find the provider that matches the service id
	var oidcProvider *conjurapi.OidcProviderResponse
	for _, provider := range data {
		if provider.ServiceID == serviceID {
			oidcProvider = &provider
			break
		}
	}
	if oidcProvider == nil {
		return nil, fmt.Errorf("OIDC provider with service ID %s not found", serviceID)
	}

	return oidcProvider, nil
}
