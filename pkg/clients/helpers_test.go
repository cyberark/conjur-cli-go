package clients

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/cyberark/conjur-api-go/conjurapi"
)

// fakeConjurClient implements ConjurClient with minimal behavior for tests.
// All methods return zero values except GetConfig and GetHttpClient which are configurable.
type fakeConjurClient struct {
	config     conjurapi.Config
	httpClient *http.Client
}

var _ ConjurClient = (*fakeConjurClient)(nil)

func (f *fakeConjurClient) Login(login string, password string) ([]byte, error) { return nil, nil }
func (f *fakeConjurClient) GetConfig() conjurapi.Config                         { return f.config }
func (f *fakeConjurClient) GetAuthenticator() conjurapi.Authenticator           { return nil }
func (f *fakeConjurClient) WhoAmI() ([]byte, error)                             { return nil, nil }
func (f *fakeConjurClient) RotateUserAPIKey(userID string) ([]byte, error)      { return nil, nil }
func (f *fakeConjurClient) RotateCurrentUserAPIKey() ([]byte, error)            { return nil, nil }
func (f *fakeConjurClient) ChangeCurrentUserPassword(newPassword string) ([]byte, error) {
	return nil, nil
}
func (f *fakeConjurClient) RotateHostAPIKey(hostID string) ([]byte, error) { return nil, nil }
func (f *fakeConjurClient) InternalAuthenticate() ([]byte, error)          { return nil, nil }
func (f *fakeConjurClient) LoadPolicy(mode conjurapi.PolicyMode, policyID string, policy io.Reader) (*conjurapi.PolicyResponse, error) {
	return nil, nil
}
func (f *fakeConjurClient) FetchPolicy(policyID string, returnJSON bool, policyTreeDepth uint, sizeLimit uint) ([]byte, error) {
	return nil, nil
}
func (f *fakeConjurClient) DryRunPolicy(mode conjurapi.PolicyMode, policyID string, policy io.Reader) (*conjurapi.DryRunPolicyResponse, error) {
	return nil, nil
}
func (f *fakeConjurClient) AddSecret(variableID string, secretValue string) error { return nil }
func (f *fakeConjurClient) RetrieveSecret(variableID string) ([]byte, error)      { return nil, nil }
func (f *fakeConjurClient) RetrieveBatchSecretsSafe(variableIDs []string) (map[string][]byte, error) {
	return nil, nil
}
func (f *fakeConjurClient) RetrieveSecretWithVersion(variableID string, version int) ([]byte, error) {
	return nil, nil
}
func (f *fakeConjurClient) CheckPermission(resourceID string, privilege string) (bool, error) {
	return false, nil
}
func (f *fakeConjurClient) CheckPermissionForRole(resourceID string, roleID string, privilege string) (bool, error) {
	return false, nil
}
func (f *fakeConjurClient) ResourceExists(resourceID string) (bool, error) { return false, nil }
func (f *fakeConjurClient) Resource(resourceID string) (map[string]interface{}, error) {
	return nil, nil
}
func (f *fakeConjurClient) Resources(filter *conjurapi.ResourceFilter) ([]map[string]interface{}, error) {
	return nil, nil
}
func (f *fakeConjurClient) ResourcesCount(filter *conjurapi.ResourceFilter) (*conjurapi.ResourcesCount, error) {
	return nil, nil
}
func (f *fakeConjurClient) PermittedRoles(resourceID, privilege string) ([]string, error) {
	return nil, nil
}
func (f *fakeConjurClient) ListOidcProviders() ([]conjurapi.OidcProvider, error) { return nil, nil }
func (f *fakeConjurClient) RefreshToken() error                                  { return nil }
func (f *fakeConjurClient) ForceRefreshToken() error                             { return nil }
func (f *fakeConjurClient) GetHttpClient() *http.Client                          { return f.httpClient }
func (f *fakeConjurClient) RoleExists(roleID string) (bool, error)               { return false, nil }
func (f *fakeConjurClient) Role(roleID string) (map[string]interface{}, error)   { return nil, nil }
func (f *fakeConjurClient) RoleMembers(roleID string) ([]map[string]interface{}, error) {
	return nil, nil
}
func (f *fakeConjurClient) RoleMembershipsAll(roleID string) (memberships []string, err error) {
	return nil, nil
}
func (f *fakeConjurClient) CreateToken(durationStr string, hostFactory string, cidrs []string, count int) ([]conjurapi.HostFactoryTokenResponse, error) {
	return nil, nil
}
func (f *fakeConjurClient) DeleteToken(token string) error { return nil }
func (f *fakeConjurClient) CreateHost(id string, token string) (conjurapi.HostFactoryHostResponse, error) {
	return conjurapi.HostFactoryHostResponse{}, nil
}
func (f *fakeConjurClient) PublicKeys(kind string, identifier string) ([]byte, error) {
	return nil, nil
}
func (f *fakeConjurClient) EnableAuthenticator(authenticatorType string, serviceID string, enabled bool) error {
	return nil
}
func (f *fakeConjurClient) Issuer(issuerID string) (issuer conjurapi.Issuer, err error) {
	return conjurapi.Issuer{}, nil
}
func (f *fakeConjurClient) Issuers() (issuers []conjurapi.Issuer, err error)     { return nil, nil }
func (f *fakeConjurClient) DeleteIssuer(issuerID string, keepSecrets bool) error { return nil }
func (f *fakeConjurClient) CreateIssuer(issuer conjurapi.Issuer) (created conjurapi.Issuer, err error) {
	return conjurapi.Issuer{}, nil
}
func (f *fakeConjurClient) UpdateIssuer(issuerID string, issuerUpdate conjurapi.IssuerUpdate) (updated conjurapi.Issuer, err error) {
	return conjurapi.Issuer{}, nil
}
func (f *fakeConjurClient) LdapCreateGroupMapping(serviceID, groupName string, roles []string) (*conjurapi.LdapGroupMappingResponse, error) {
	return nil, nil
}
func (f *fakeConjurClient) LdapShowGroupMapping(serviceID, groupName string) (*conjurapi.LdapGroupMappingResponse, error) {
	return nil, nil
}
func (f *fakeConjurClient) LdapListGroupMappings(serviceID string) (*conjurapi.LdapMappingListResponse, error) {
	return nil, nil
}
func (f *fakeConjurClient) LdapDeleteGroupMapping(serviceID, groupName string) error { return nil }
func (f *fakeConjurClient) LdapCreateUserMapping(serviceID, username string, roles []string) (*conjurapi.LdapUserMappingResponse, error) {
	return nil, nil
}
func (f *fakeConjurClient) LdapShowUserMapping(serviceID, username string) (*conjurapi.LdapUserMappingResponse, error) {
	return nil, nil
}
func (f *fakeConjurClient) LdapListUserMappings(serviceID string) (*conjurapi.LdapMappingListResponse, error) {
	return nil, nil
}
func (f *fakeConjurClient) LdapDeleteUserMapping(serviceID, username string) error { return nil }

// rewriteTransport redirects requests to a test server, preserving path and query.
type rewriteTransport struct {
	rt        http.RoundTripper
	targetURL *url.URL
}

func (r *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := req.Clone(req.Context())
	req2.URL.Scheme = r.targetURL.Scheme
	req2.URL.Host = r.targetURL.Host
	return r.rt.RoundTrip(req2)
}

// newTestClient returns an HTTP client that routes all requests to srv.
func newTestClient(srv *httptest.Server) *http.Client {
	u, _ := url.Parse(srv.URL)
	return &http.Client{Transport: &rewriteTransport{rt: http.DefaultTransport, targetURL: u}}
}
