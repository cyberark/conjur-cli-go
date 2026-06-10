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

type CloudURL struct {
	Tenant string
	Suffix string
}

// tenantDiscoveryResp models the /api/public/tenant-discovery response.
type tenantDiscoveryResp struct {
	TenantID string                   `json:"tenant_id"`
	Services []tenantDiscoveryService `json:"services"`
}

type tenantDiscoveryService struct {
	ServiceName string                           `json:"service_name"`
	Region      string                           `json:"region"`
	Endpoints   []tenantDiscoveryServiceEndpoint `json:"endpoints"`
}

type tenantDiscoveryServiceEndpoint struct {
	IsActive bool   `json:"is_active"`
	Type     string `json:"type"`
	UI       string `json:"ui"`
	API      string `json:"api"`
}

const (
	// tenantDiscoveryIdentityService is the service name for identity administration
	// in the /api/public/tenant-discovery response.
	tenantDiscoveryIdentityService = "identity_administration"
	tenantDiscoveryMainEndpoint    = "main"
)

func CloudLogin(conjurClient ConjurClient, username string, password string, insecureLogin bool) (ConjurClient, error) {
	if strings.HasPrefix(username, "host/") {
		return cloudHostLogin(conjurClient, username, password)
	}
	return cloudIdentityLogin(conjurClient, username, password, insecureLogin)
}

func cloudHostLogin(conjurClient ConjurClient, username string, password string) (ConjurClient, error) {
	config := conjurClient.GetConfig()

	username, password, err := prompts.MaybeAskForCredentials(username, password)
	if err != nil {
		return nil, err
	}

	return conjurapi.NewClientFromCloudHost(config, username, password, TelemetryData)
}

func cloudIdentityLogin(client ConjurClient, username, password string, insecureLogin bool) (ConjurClient, error) {
	identityURL, tenantID, err := identityURL(client)
	if err != nil {
		return nil, err
	}
	username, err = prompts.MaybeAskForUsername(username)
	if err != nil {
		return nil, err
	}
	authToken, err := tokenFromIdentity(client, identityURL, tenantID, username, password, insecureLogin)
	if err != nil {
		return nil, err
	}

	client, err = conjurapi.NewClientFromOidcToken(client.GetConfig(), authToken, TelemetryData)
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
				"Your credentials are valid, failed to authenticate with Idira Secrets Manager, SaaS. "+
					"Please enable debug mode for more details, try again later or contact your administrator"+
					" for assistance: %s",
				http.StatusText(errorCode),
			)
		}
		return nil, errors.New(
			"OIDC login was successful, but your access was denied by Idira Secrets Manager, SaaS. " +
				"Please enable debug mode for more details, verify your account roles, and try again " +
				"later or contact your administrator for assistance",
		)
	}

	return client, nil
}

func tokenFromIdentity(client ConjurClient, url string, tenantID string, username string, password string, insecureLogin bool) (string, error) {
	ia := NewIdentityAuthenticator(client, url, tenantID, insecureLogin)
	return ia.GetToken(username, password)
}

// identityURL returns the identity (idaptive) UI URL and tenant ID for the
// configured appliance by querying the platform-discovery service.
func identityURL(client ConjurClient) (string, string, error) {
	ccURL, err := ParseCloudURL(client.GetConfig().ApplianceURL)
	if err != nil {
		return "", "", err
	}

	result, err := fetchTenantDiscovery(client.GetHttpClient(), tenantDiscoveryEndpointURL(ccURL))
	if err != nil {
		return "", "", err
	}

	if result.TenantID == "" {
		return "", "", fmt.Errorf("tenant_id not found in tenant discovery response")
	}

	identityAdminURL, err := findIdentityAdminURL(result)
	if err != nil {
		return "", "", err
	}

	return identityAdminURL, result.TenantID, nil
}

// tenantDiscoveryEndpointURL builds the platform-discovery request URL for the
// given CloudURL. The tenant is passed as a query parameter so that the shared
// platform-discovery host can serve all tenants on the same domain suffix.
func tenantDiscoveryEndpointURL(ccURL *CloudURL) string {
	return fmt.Sprintf(
		"https://platform-discovery%s/api/public/tenant-discovery?bySubdomain=%s&allEndpoints=true",
		ccURL.Suffix, ccURL.Tenant,
	)
}

func fetchTenantDiscovery(httpClient *http.Client, url string) (*tenantDiscoveryResp, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("tenant discovery request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read tenant discovery response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tenant discovery request returned HTTP %d", resp.StatusCode)
	}

	var result tenantDiscoveryResp
	if err = json.Unmarshal(respBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to parse tenant discovery response: %w", err)
	}

	return &result, nil
}

// findIdentityAdminURL returns the UI URL of the identity_administration service's
// main endpoint from the tenant discovery response.
func findIdentityAdminURL(result *tenantDiscoveryResp) (string, error) {
	for _, svc := range result.Services {
		if svc.ServiceName != tenantDiscoveryIdentityService {
			continue
		}
		for _, ep := range svc.Endpoints {
			if ep.Type == tenantDiscoveryMainEndpoint && ep.UI != "" {
				return ep.UI, nil
			}
		}
		return "", fmt.Errorf("main endpoint not found for identity_administration service in tenant discovery response")
	}
	return "", fmt.Errorf("identity_administration service not found in tenant discovery response")
}

func ParseCloudURL(url string) (*CloudURL, error) {
	var res CloudURL

	matches := regexp.MustCompile("^https://([^.]+)(.*)(/api)$").FindStringSubmatch(url)
	if len(matches) == 0 {
		return nil, fmt.Errorf(
			"invalid Idira Secrets Manager, SaaS URL: " +
				"expected format https://<tenant>.secretsmgr.cyberark.cloud/api",
		)
	}
	res.Tenant = strings.TrimSuffix(matches[1], "-secretsmanager")
	res.Suffix = strings.TrimPrefix(matches[2], ".secretsmgr")

	if !strings.Contains(url, "-secretsmanager") && !strings.Contains(url, ".secretsmgr") {
		return &res, fmt.Errorf(
			"invalid Idira Secrets Manager, SaaS URL: "+
				"expected format https://<tenant>.secretsmgr.cyberark.cloud/api "+
				"Did you mean? \"https://%s.secretsmgr.cyberark.cloud/api\"", res.Tenant,
		)
	}
	for _, suf := range conjurapi.ConjurCloudSuffixes {
		if strings.HasSuffix(res.Suffix, suf) {
			return &res, nil
		}
	}
	return &res, fmt.Errorf("invalid Idira Secrets Manager, SaaS URL: "+
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
