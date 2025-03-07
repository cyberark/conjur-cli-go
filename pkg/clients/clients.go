package clients

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cyberark/conjur-api-go/conjurapi"

	"github.com/spf13/cobra"
)

// ConjurClient is an interface that represents a Conjur client
type ConjurClient interface {
	Login(login string, password string) ([]byte, error)
	GetConfig() conjurapi.Config
	GetAuthenticator() conjurapi.Authenticator
	WhoAmI() ([]byte, error)
	RotateUserAPIKey(userID string) ([]byte, error)
	RotateCurrentUserAPIKey() ([]byte, error)
	ChangeCurrentUserPassword(newPassword string) ([]byte, error)
	RotateHostAPIKey(hostID string) ([]byte, error)
	InternalAuthenticate() ([]byte, error)
	LoadPolicy(mode conjurapi.PolicyMode, policyID string, policy io.Reader) (*conjurapi.PolicyResponse, error)
	FetchPolicy(policyID string, returnJSON bool, policyTreeDepth uint, sizeLimit uint) ([]byte, error)
	DryRunPolicy(mode conjurapi.PolicyMode, policyID string, policy io.Reader) (*conjurapi.DryRunPolicyResponse, error)
	AddSecret(variableID string, secretValue string) error
	RetrieveSecret(variableID string) ([]byte, error)
	RetrieveBatchSecretsSafe(variableIDs []string) (map[string][]byte, error)
	RetrieveSecretWithVersion(variableID string, version int) ([]byte, error)
	CheckPermission(resourceID string, privilege string) (bool, error)
	CheckPermissionForRole(resourceID string, roleID string, privilege string) (bool, error)
	ResourceExists(resourceID string) (bool, error)
	Resource(resourceID string) (resource map[string]interface{}, err error)
	Resources(filter *conjurapi.ResourceFilter) ([]map[string]interface{}, error)
	ResourcesCount(filter *conjurapi.ResourceFilter) (*conjurapi.ResourcesCount, error)
	PermittedRoles(resourceID, privilege string) ([]string, error)
	ListOidcProviders() ([]conjurapi.OidcProvider, error)
	RefreshToken() error
	ForceRefreshToken() error
	GetHttpClient() *http.Client
	RoleExists(roleID string) (bool, error)
	Role(roleID string) (role map[string]interface{}, err error)
	RoleMembers(roleID string) (members []map[string]interface{}, err error)
	RoleMembershipsAll(roleID string) (memberships []string, err error)
	CreateToken(durationStr string, hostFactory string, cidrs []string, count int) ([]conjurapi.HostFactoryTokenResponse, error)
	DeleteToken(token string) error
	CreateHost(id string, token string) (conjurapi.HostFactoryHostResponse, error)
	PublicKeys(kind string, identifier string) ([]byte, error)
}

// LoadAndValidateConjurConfig loads and validate Conjur configuration
func LoadAndValidateConjurConfig(timeout time.Duration) (conjurapi.Config, error) {
	// TODO: extract this common code for gathering configuring into a seperate package
	// Some of the code is in conjur-api-go and needs to be made configurable so that you can pass a custom path to .conjurrc

	config, err := conjurapi.LoadConfig()
	if err != nil {
		return config, err
	}

	if timeout > 0 {
		// do not overwrite the value set in the env or .conjurrc file
		config.HTTPTimeout = int(timeout.Seconds())
	}
	err = config.Validate()

	return config, err
}

// AuthenticatedConjurClientForCommand attempts to get an authenticated Conjur client by iterating through
// configuration, environment variables and then ultimately falling back on prompting the user for credentials.
func AuthenticatedConjurClientForCommand(cmd *cobra.Command) (ConjurClient, error) {
	var err error

	debug, err := cmd.Flags().GetBool("debug")
	if err != nil {
		return nil, err
	}

	// TODO: This is called multiple time because each operation potentially uses a new HTTP client bound to the
	// temporary Conjur client being created at that point in time. We should really not be creating so many Conjur clients
	// we should just have one then the rest is an attempt to get an authenticator
	decorateConjurClient := func(client ConjurClient) {
		MaybeDebugLoggingForClient(debug, cmd, client)
	}

	timeout, err := GetTimeout(cmd)
	if err != nil {
		return nil, err
	}

	config, err := LoadAndValidateConjurConfig(timeout)
	if err != nil {
		return nil, err
	}

	var client ConjurClient
	client, err = conjurapi.NewClientFromEnvironment(config)
	if err != nil {
		return nil, err
	}
	decorateConjurClient(client)

	if client.GetAuthenticator() == nil {
		client, err = conjurapi.NewClient(config)
		if err != nil {
			return nil, err
		}
		decorateConjurClient(client)

		if config.AuthnType == "" || config.AuthnType == "authn" || config.AuthnType == "ldap" {
			client, err = Login(client)
		} else if config.AuthnType == "oidc" {
			client, err = OidcLogin(client, "", "")
		} else if config.AuthnType == "jwt" {
			// Will use the token in the config
		} else {
			return nil, fmt.Errorf("unsupported authentication type: %s", config.AuthnType)
		}

		if err != nil {
			return nil, err
		}
		decorateConjurClient(client)
	}

	return client, nil
}

// GetTimeout extracts the timeout from the command flags only if explicitly set
func GetTimeout(cmd *cobra.Command) (timeout time.Duration, err error) {
	if cmd.Flags().Changed("timeout") {
		// do not use the cobra default value in case it was not explicitly provided
		// otherwise it would overwrite the value set in the env or .conjurrc file
		timeout, err = cmd.Flags().GetDuration("timeout")
		if err != nil {
			return 0, err
		}
	}
	return timeout, nil
}
