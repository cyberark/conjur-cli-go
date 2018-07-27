package action

// VariableClient contains the Conjur API methods required to
// implement the variable actions.
type VariableClient interface {
	RetrieveSecret(name string) ([]byte, error)
	AddSecret(name, value string) error
}

// Variable contains the information required by the variable actions.
type Variable struct {
	Name string
}

// Value uses the Conjur API to retrieve a secret.
func (v Variable) Value(client VariableClient) ([]byte, error) {
	return client.RetrieveSecret(v.Name)
}

// ValuesAdd uses the Conjur API to add a value for a secret.
func (v Variable) ValuesAdd(client VariableClient, value string) error {
	return client.AddSecret(v.Name, value)
}
