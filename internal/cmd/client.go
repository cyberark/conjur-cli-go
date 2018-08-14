package cmd

import (
	"io"

	"github.com/cyberark/conjur-api-go/conjurapi"
)

// PolicyClient specifies the Conjur API methods required to implement the policy actions.
type PolicyClient interface {
	LoadPolicy(mode conjurapi.PolicyMode, policyID string, policy io.Reader) (*conjurapi.PolicyResponse, error)
}

// VariableClient contains the Conjur API methods required to
// implement the variable actions.
type VariableClient interface {
	RetrieveSecret(id string) ([]byte, error)
	AddSecret(id, value string) error
}

// ConjurClient is the composition of all the Conjur API interfaces used by the individual action clients.
type ConjurClient interface {
	PolicyClient
	VariableClient
}
