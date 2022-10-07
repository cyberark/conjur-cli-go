package authn

import (
	"fmt"

	"github.com/cyberark/conjur-api-go/conjurapi"
)

func OidcLogin(config conjurapi.Config, decorateClient func(*conjurapi.Client)) (*conjurapi.Client, error) {
	conjurClient, err := conjurapi.NewClient(config)
	if err != nil {
		return nil, err
	}

	decorateClient(conjurClient)

	oidcProvider, err := getOidcProviderInfo(conjurClient, config.ServiceID)
	if err != nil {
		return nil, err
	}

	resopnse, err := HandleOpenIDFlow(oidcProvider.RedirectURI)
	if err != nil {
		return nil, err
	}

	//TODO: What to store?
	// err = StoreCredentials(config, config.ServiceID, jwtContent)
	// if err != nil {
	// 	return nil, err
	// }

	conjurClient, err = conjurapi.NewClientFromOidcCode(config, response.code, response.state)
	decorateClient(conjurClient)

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
