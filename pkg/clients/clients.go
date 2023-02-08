package clients

import (
	"fmt"
	"io"
	"net/http"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-cli-go/pkg/prompts"

	"github.com/spf13/cobra"
)

// ConjurClient is an interface that represents a Conjur client
type ConjurClient interface {
	Login(login string, password string) ([]byte, error)
	GetConfig() conjurapi.Config
	GetAuthenticator() conjurapi.Authenticator
	WhoAmI() ([]byte, error)
	RotateUserAPIKey(userID string) ([]byte, error)
	ChangeCurrentUserPassword(newPassword string) ([]byte, error)
	RotateHostAPIKey(hostID string) ([]byte, error)
	InternalAuthenticate() ([]byte, error)
	LoadPolicy(mode conjurapi.PolicyMode, policyID string, policy io.Reader) (*conjurapi.PolicyResponse, error)
	AddSecret(variableID string, secretValue string) error
	RetrieveSecret(variableID string) ([]byte, error)
	CheckPermission(resourceID string, privilege string) (bool, error)
	CheckPermissionForRole(resourceID string, roleID string, privilege string) (bool, error)
	ResourceExists(resourceID string) (bool, error)
	Resource(resourceID string) (resource map[string]interface{}, err error)
	Resources(filter *conjurapi.ResourceFilter) ([]map[string]interface{}, error)
	PermittedRoles(resourceID, privilege string) ([]string, error)
	ListOidcProviders() ([]conjurapi.OidcProvider, error)
	RefreshToken() error
	ForceRefreshToken() error
	GetHttpClient() *http.Client
	RoleExists(roleID string) (bool, error)
	Role(roleID string) (role map[string]interface{}, err error)
	RoleMembers(roleID string) (members []map[string]interface{}, err error)
	RoleMemberships(roleID string) (memberships []map[string]interface{}, err error)
	CreateToken(durationStr string, hostFactory string, cidrs []string, count int) ([]conjurapi.HostFactoryTokenResponse, error)
	DeleteToken(token string) error
	CreateHost(id string, token string) (conjurapi.HostFactoryHostResponse, error)
	PublicKeys(kind string, identifier string) ([]byte, error)
}

// LoadAndValidateConjurConfig loads and validate Conjur configuration
func LoadAndValidateConjurConfig() (conjurapi.Config, error) {
	// TODO: extract this common code for gathering configuring into a seperate package
	// Some of the code is in conjur-api-go and needs to be made configurable so that you can pass a custom path to .conjurrc

	config, err := conjurapi.LoadConfig()
	if err != nil {
		return config, err
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

	config, err := LoadAndValidateConjurConfig()
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
			decoratePrompt := prompts.PromptDecoratorForCommand(cmd)
			client, err = Login(client, decoratePrompt)
		} else if config.AuthnType == "oidc" {
			client, err = OidcLogin(client, "", "")
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
