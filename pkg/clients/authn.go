package clients

import (
	"errors"
	"fmt"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-api-go/conjurapi/authn"
	"github.com/cyberark/conjur-cli-go/pkg/prompts"
	"github.com/manifoldco/promptui"
)

type oidcConjurClient interface {
	ListOidcProviders() ([]conjurapi.OidcProvider, error)
}

// Login attempts to login to Conjur and prompts the user for credentials interactively
func Login(conjurClient ConjurClient, decoratePrompt func(prompt *promptui.Prompt) *promptui.Prompt) (ConjurClient, error) {
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

	return authenticatePair, nil
}

// oidcLogin attempts to login to Conjur using the OIDC flow
func oidcLogin(conjurClient ConjurClient, oidcPromptHandler func(string) error) (ConjurClient, error) {
	config := conjurClient.GetConfig()

	oidcProvider, err := getOidcProviderInfo(conjurClient, config.ServiceID)
	if err != nil {
		return nil, err
	}

	code, err := handleOpenIDFlow(oidcProvider.RedirectURI, generateState, oidcPromptHandler)
	if err != nil {
		return nil, err
	}

	conjurClient, err = conjurapi.NewClientFromOidcCode(config, code, oidcProvider.Nonce, oidcProvider.CodeVerifier)
	if err != nil {
		return nil, err
	}

	// Refreshes the access token and caches it locally
	err = conjurClient.ForceRefreshToken()
	if err != nil {
		return nil, errors.New("Unable to authenticate with Conjur. Please check your credentials.")
	}

	return conjurClient, nil
}

func getOidcProviderInfo(conjurClient oidcConjurClient, serviceID string) (*conjurapi.OidcProvider, error) {
	data, err := conjurClient.ListOidcProviders()
	if err != nil {
		return nil, err
	}

	// Find the provider that matches the service id
	var oidcProvider *conjurapi.OidcProvider
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
