package clients

import (
	"fmt"

	"github.com/cyberark/conjur-api-go/conjurapi"
)

func OidcLogin(conjurClient *conjurapi.Client) (*conjurapi.Client, error) {
	oidcProvider, err := getOidcProviderInfo(conjurClient, conjurClient.GetConfig().ServiceID)
	if err != nil {
		return nil, err
	}

	code, err := handleOpenIDFlow(oidcProvider.RedirectURI)
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
