package clients

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-cli-go/pkg/prompts"
	"io"
	"regexp"
	"strings"
)

func CloudLogin(conjurClient ConjurClient, username string, password string) (ConjurClient, error) {
	if strings.HasPrefix(username, "/host/") {
		return cloudHostLogin(conjurClient, username, password)
	}
	return cloudIdentityLogin(conjurClient, username, password)
}

func cloudHostLogin(conjurClient ConjurClient, username string, password string) (ConjurClient, error) {
	authenticatePair, err := LoginWithPromptFallback(conjurClient, username, password)
	if err != nil {
		return nil, err
	}
	return conjurapi.NewClientFromKey(conjurClient.GetConfig(), *authenticatePair)
}

func cloudIdentityLogin(client ConjurClient, username, password string) (ConjurClient, error) {
	identityURL, err := identityURL(client)
	if err != nil {
		return nil, err
	}
	username, err = prompts.MaybeAskForUsername(username)
	if err != nil {
		return nil, err
	}
	authToken, err := tokenFromIdentity(client, identityURL, username, password)
	if err != nil {
		return nil, err
	}

	client, err = conjurapi.NewClientFromOidcToken(client.GetConfig(), authToken)
	if err != nil {
		return nil, err
	}

	// Refreshes the access token and caches it locally
	err = client.ForceRefreshToken()

	if err != nil {
		// 401 Unauthorized might be returned when the user is not provisioned in Conjur Cloud, but only exists in Identity
		if strings.Contains(err.Error(), "401 Unauthorized") {
            return nil, errors.New("Please check your credentials. Consider adding Conjur Cloud Admin and Conjur Cloud User roles to your account")
		}
		return nil, errors.New("Unable to authenticate with Secrets Manager SaaS. Please check your credentials.")
	}

	return client, nil
}

func tokenFromIdentity(client ConjurClient, url string, username string, password string) (string, error) {
	ia := NewIdentityAuthenticator(client, url)
	return ia.GetToken(username, password)
}

type endpointResp struct {
	Fqdn         string `json:"fqdn"`
	TenantMgmtId string `json:"tenant_mgmt_id"`
}

func identityURL(client ConjurClient) (string, error) {
	tenantURL := client.GetConfig().ApplianceURL
	matches := regexp.MustCompile("^https://([^.]+)(.*)(/api)$").FindStringSubmatch(tenantURL)
	if len(matches) == 0 {
		return "", fmt.Errorf("invalid Secrets Manager SaaS URL: %s", tenantURL)
	}
	tenant := strings.TrimSuffix(matches[1], "-secretsmanager")
	rest := strings.TrimPrefix(matches[2], ".secretsmgr")
	endpoint := fmt.Sprintf("https://%s%s/shell/api/endpoint/%s", tenant, rest, tenant)
	response, err := client.GetHttpClient().Get(endpoint)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	var result endpointResp
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("https://%s", result.Fqdn), nil
}
