package cmd

import (
	"io"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-api-go/conjurapi/authn"
)

// AuthnClient specifies the Conjur API methods required to implement the authentication actions.
type AuthnClient interface {
	GetConfig() conjurapi.Config
	GetAuthToken() authn.AuthnToken
	RefreshToken() error
	Login(username, password string) ([]byte, error)
}

// PolicyClient specifies the Conjur API methods required to implement the policy actions.
type PolicyClient interface {
	LoadPolicy(mode conjurapi.PolicyMode, policyID string, policy io.Reader) (*conjurapi.PolicyResponse, error)
}

// ResourceClient specifies the Conjur API methods used by resource actions.
type ResourceClient interface {
	Resource(resourceID string) (map[string]interface{}, error)
	Resources(filter *conjurapi.ResourceFilter) ([]map[string]interface{}, error)
}

// VariableClient contains the Conjur API methods required to
// implement the variable actions.
type VariableClient interface {
	RetrieveSecret(id string) ([]byte, error)
	AddSecret(id, value string) error
}

// ConjurClient is the composition of all the Conjur API interfaces used by the individual action clients.
type ConjurClient interface {
	AuthnClient
	PolicyClient
	ResourceClient
	VariableClient
}
