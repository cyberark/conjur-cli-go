package clients

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-api-go/conjurapi/response"
	"github.com/cyberark/conjur-cli-go/pkg/prompts"
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
	identityURL, tenantID, err := identityURL(client)
	if err != nil {
		return nil, err
	}
	username, err = prompts.MaybeAskForUsername(username)
	if err != nil {
		return nil, err
	}
	authToken, err := tokenFromIdentity(client, identityURL, tenantID, username, password)
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
		errorCode := extractHTTPStatusCode(err)
		// 401 Unauthorized might be returned when the user is not provisioned in Conjur Cloud, but only exists in Identity
		if errorCode == http.StatusUnauthorized {
			return nil, errors.New(
				"Your credentials are valid, but your account may be missing required roles. " +
					"Please ensure your account has the admin or user roles, try again later " +
					"or contact your administrator to update your configuration",
			)
		}
		if errorCode >= 400 {
			return nil, fmt.Errorf(
				"Your credentials are valid, failed to authenticate with Secrets Manager SaaS. "+
					"Please enable debug mode for more details, try again later or contact your administrator"+
					" for assistance: %s",
				http.StatusText(errorCode),
			)
		}
		return nil, errors.New(
			"OIDC login was successful, but your access was denied by Secrets Manager SaaS. " +
				"Please enable debug mode for more details, verify your account roles, and try again " +
				"later or contact your administrator for assistance",
		)
	}

	return client, nil
}

func tokenFromIdentity(client ConjurClient, url string, tenantID string, username string, password string) (string, error) {
	ia := NewIdentityAuthenticator(client, url, tenantID)
	return ia.GetToken(username, password)
}

type endpointResp struct {
	Fqdn         string `json:"fqdn"`
	TenantMgmtId string `json:"tenant_mgmt_id"`
}

func identityURL(client ConjurClient) (string, string, error) {
	tenantURL := client.GetConfig().ApplianceURL
	ccURL, err := ParseCloudURL(tenantURL)
	if err != nil {
		return "", "", err
	}
	endpoint := fmt.Sprintf("https://%s%s/shell/api/endpoint/%s", ccURL.Tenant, ccURL.Suffix, ccURL.Tenant)
	resp, err := client.GetHttpClient().Get(endpoint)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	var result endpointResp
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return "", "", err
	}
	return fmt.Sprintf("https://%s", result.Fqdn), ccURL.Tenant, nil
}

type CloudURL struct {
	Tenant string
	Suffix string
}

func ParseCloudURL(url string) (*CloudURL, error) {
	var res CloudURL

	matches := regexp.MustCompile("^https://([^.]+)(.*)(/api)$").FindStringSubmatch(url)
	if len(matches) == 0 {
		return nil, fmt.Errorf(
			"invalid Secrets Manager SaaS URL: " +
				"expected format https://<tenant>.secretsmgr.cyberark.cloud/api",
		)
	}
	res.Tenant = strings.TrimSuffix(matches[1], "-secretsmanager")
	res.Suffix = strings.TrimPrefix(matches[2], ".secretsmgr")

	if !strings.Contains(url, "-secretsmanager") && !strings.Contains(url, ".secretsmgr") {
		return &res, fmt.Errorf(
			"invalid Secrets Manager SaaS URL: "+
				"expected format https://<tenant>.secretsmgr.cyberark.cloud/api "+
				"Did you mean? \"https://%s.secretsmgr.cyberark.cloud/api\"", res.Tenant,
		)
	}
	for _, suf := range conjurapi.ConjurCloudSuffixes {
		if strings.HasSuffix(res.Suffix, suf) {
			return &res, nil
		}
	}
	return &res, fmt.Errorf("invalid Secrets Manager SaaS URL: "+
		"expected format https://<tenant>.secretsmgr.cyberark.cloud/api "+
		"unrecognized domain suffix \"%s\"", res.Suffix)
}

func extractHTTPStatusCode(err error) int {
	var cerr *response.ConjurError
	if errors.As(err, &cerr) {
		return cerr.Code
	}
	return 0
}
